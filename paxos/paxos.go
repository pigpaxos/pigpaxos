package paxos

import (
	"github.com/pigpaxos/pigpaxos/hlc"
	"github.com/pigpaxos/pigpaxos/log"
	"strconv"
	"sync"
	"time"

	"github.com/pigpaxos/pigpaxos"
)

// this is the difference we allow between max seen slot and executed slot
// if executed + ExecuteSlack < MaxSlot then we want try to recover the slots
// as this is likely the case of state machine getting stuck because of network/communication failures
const ExecuteSlack = 10

// entry in log
type entry struct {
	ballot    paxi.Ballot
	command   paxi.Command
	commit    bool
	request   *paxi.Request
	quorum    *paxi.Quorum
	timestamp time.Time
}

// Paxos instance
type Paxos struct {
	paxi.Node

	config []paxi.ID

	log     map[int]*entry // log ordered by slot
	execute int            // next execute slot number
	active  bool           // active leader
	ballot  paxi.Ballot    // highest ballot number
	slot    int            // highest slot number

	p3PendingBallot   paxi.Ballot
	p3pendingSlots    []int
	lastP3Time        int64
	lastCleanupMarker int
	globalExecuted    int             // executed by all nodes. Need for log cleanup
	executeByNode     map[paxi.ID]int // leader's knowledge of other nodes execute counter. Need for log cleanup

	quorum   *paxi.Quorum    // phase 1 quorum
	requests []*paxi.Request // phase 1 pending requests

	Q1              func(*paxi.Quorum) bool
	Q2              func(*paxi.Quorum) bool
	ReplyWhenCommit bool

	// Locks
	logLck			sync.RWMutex
	p3Lock			sync.RWMutex
	markerLock		sync.RWMutex
}

// NewPaxos creates new paxos instance
func NewPaxos(n paxi.Node, options ...func(*Paxos)) *Paxos {
	p := &Paxos{
		Node:            n,
		log:             make(map[int]*entry, paxi.GetConfig().BufferSize),
		slot:            -1,
		quorum:          paxi.NewQuorum(),
		requests:        make([]*paxi.Request, 0),
		Q1:              func(q *paxi.Quorum) bool { return q.Majority() },
		Q2:              func(q *paxi.Quorum) bool { return q.Majority() },
		ReplyWhenCommit: false,
	}

	for _, id := range paxi.GetConfig().IDs() {
		p.executeByNode[id] = 0
	}

	for _, opt := range options {
		opt(p)
	}

	return p
}

// IsLeader indecates if this node is current leader
func (p *Paxos) IsLeader() bool {
	return p.active || p.ballot.ID() == p.ID()
}

// Leader returns leader id of the current ballot
func (p *Paxos) Leader() paxi.ID {
	return p.ballot.ID()
}

// Ballot returns current ballot
func (p *Paxos) Ballot() paxi.Ballot {
	return p.ballot
}

// SetActive sets current paxos instance as active leader
func (p *Paxos) SetActive(active bool) {
	p.active = active
}

// SetBallot sets a new ballot number
func (p *Paxos) SetBallot(b paxi.Ballot) {
	p.ballot = b
}

// forceful sync of P3 messages if no progress was done in past 10 ms and no P3 msg was piggybacked
func (p *Paxos) P3Sync(tnow int64) {
	p.logLck.RLock()
	defer p.logLck.RUnlock()
	log.Debugf("Syncing P3 (tnow=%d)", tnow)
	if tnow - 10 > p.lastP3Time && p.lastP3Time > 0 && p.p3PendingBallot > 0 {
		p.p3Lock.Lock()
		p.Broadcast(P3{
			Ballot:  p.p3PendingBallot,
			Slot:    p.p3pendingSlots,
		})
		p.lastP3Time = tnow
		p.p3pendingSlots = make([]int, 0, 100)
		p.p3Lock.Unlock()
	}
}

func (p *Paxos) CheckNeedForRecovery() {
	p.logLck.RLock()
	defer p.logLck.RUnlock()
	// check if we state machine appears stuck and needs to recover some slots due to missing P3s or P2bs or both
	if p.execute + ExecuteSlack < p.slot {
		recoverSlots := make([]int, 0)
		for slot := p.execute; slot < p.slot - ExecuteSlack; slot++ {
			e, exists := p.log[slot]
			if !exists || !e.commit {
				recoverSlots = append(recoverSlots, slot)
			}
		}
		p.sendRecoverRequest(p.Ballot(), recoverSlots)
	}
}

func (p *Paxos) UpdateLastExecutedByNode(id paxi.ID, lastExecute int) {
	p.markerLock.Lock()
	defer p.markerLock.Unlock()
	p.executeByNode[id] = lastExecute
}

func (p *Paxos) GetGlobalExecuteMarker() int {
	if p.IsLeader() {
		marker := p.execute
		for _, c := range p.executeByNode {
			if c < marker {
				marker = c
			}
		}
		p.globalExecuted = marker
	}
	return p.globalExecuted
}

func (p *Paxos) CleanupLog() {
	p.markerLock.Lock()
	marker := p.GetGlobalExecuteMarker()
	log.Debugf("Replica %v log cleanup. lastCleanupMarker: %d, safeCleanUpMarker: %d", p.ID(), p.lastCleanupMarker, marker)
	p.markerLock.Unlock()

	p.logLck.Lock()
	defer p.logLck.Unlock()
	for i := p.lastCleanupMarker; i < marker; i++ {
		delete(p.log, i)
	}
	p.lastCleanupMarker = marker
}

// HandleRequest handles request and start phase 1 or phase 2
func (p *Paxos) HandleRequest(r paxi.Request) {
	// log.Debugf("Replica %s received %v\n", p.ID(), r)
	if !p.active {
		p.requests = append(p.requests, &r)
		// current phase 1 pending
		if p.ballot.ID() != p.ID() {
			p.P1a()
		}
	} else {
		p.P2a(&r)
	}
}

// P1a starts phase 1 prepare
func (p *Paxos) P1a() {
	if p.active {
		return
	}
	p.ballot.Next(p.ID())
	p.quorum.Reset()
	p.quorum.ACK(p.ID())
	p.Broadcast(P1a{Ballot: p.ballot})
}

// P2a starts phase 2 accept
func (p *Paxos) P2a(r *paxi.Request) {
	p.logLck.Lock()
	defer p.logLck.Unlock()

	p.slot++
	p.log[p.slot] = &entry{
		ballot:    p.ballot,
		command:   r.Command,
		request:   r,
		quorum:    paxi.NewQuorum(),
		timestamp: time.Now(),
	}
	p.log[p.slot].quorum.ACK(p.ID())
	m := P2a{
		Ballot: 			p.ballot,
		Slot:    			p.slot,
		GlobalExecutedSlot: p.globalExecuted,
		Command: 			r.Command,
	}

	p.p3Lock.Lock()
	if p.p3PendingBallot > 0 {
		m.P3msg = P3{Ballot: p.p3PendingBallot, Slot: p.p3pendingSlots}
		p.p3pendingSlots = make([]int, 0, 100)
		p.lastP3Time = hlc.CurrentTimeInMS()
	}
	p.p3Lock.Unlock()

	if paxi.GetConfig().Thrifty {
		p.MulticastQuorum(paxi.GetConfig().N()/2+1, m)
	} else {
		p.Broadcast(m)
	}
}


// HandleP1a handles P1a message
func (p *Paxos) HandleP1a(m P1a) {
	log.Debugf("Replica %s ===[%v]===>>> Replica %s\n", m.Ballot.ID(), m, p.ID())

	// new leader
	if m.Ballot > p.ballot {
		p.ballot = m.Ballot
		p.active = false
		// TODO use BackOff time or forward
		// forward pending requests to new leader
		p.forward()
		// if len(p.requests) > 0 {
		// 	defer p.P1a()
		// }
	}
	p.logLck.RLock()
	defer p.logLck.RUnlock()
	l := make(map[int]CommandBallot)
	for s := p.execute; s <= p.slot; s++ {
		if p.log[s] == nil || p.log[s].commit {
			continue
		}
		l[s] = CommandBallot{p.log[s].command, p.log[s].ballot}
	}

	p.Send(m.Ballot.ID(), P1b{
		Ballot: p.ballot,
		ID:     p.ID(),
		Log:    l,
	})
}

func (p *Paxos) update(scb map[int]CommandBallot) {
	p.logLck.Lock()
	defer p.logLck.Unlock()
	for s, cb := range scb {
		p.slot = paxi.Max(p.slot, s)
		if e, exists := p.log[s]; exists {
			if !e.commit && cb.Ballot > e.ballot {
				e.ballot = cb.Ballot
				e.command = cb.Command
			}
		} else {
			p.log[s] = &entry{
				ballot:  cb.Ballot,
				command: cb.Command,
				commit:  false,
			}
		}
	}
}

// HandleP1b handles P1b message
func (p *Paxos) HandleP1b(m P1b) {
	// old message
	if m.Ballot < p.ballot || p.active {
		// log.Debugf("Replica %s ignores old message [%v]\n", p.ID(), m)
		return
	}

	log.Debugf("Replica %s ===[%v]===>>> Replica %s\n", m.ID, m, p.ID())

	p.update(m.Log)

	// reject message
	if m.Ballot > p.ballot {
		p.ballot = m.Ballot
		p.active = false // not necessary
		// forward pending requests to new leader
		p.forward()
		// p.P1a()
	}

	// ack message
	if m.Ballot.ID() == p.ID() && m.Ballot == p.ballot {
		p.quorum.ACK(m.ID)
		if p.Q1(p.quorum) {
			p.active = true
			p.p3PendingBallot = p.ballot
			// propose any uncommitted entries
			for i := p.execute; i <= p.slot; i++ {
				// TODO nil gap?
				p.logLck.RLock()
				logEntry := p.log[i]
				p.logLck.RUnlock()
				if logEntry == nil || logEntry.commit {
					continue
				}
				logEntry.ballot = p.ballot
				logEntry.quorum = paxi.NewQuorum()
				logEntry.quorum.ACK(p.ID())
				p.Broadcast(P2a{
					Ballot:  			p.ballot,
					Slot:    			i,
					GlobalExecutedSlot: p.globalExecuted,
					Command: 			logEntry.command,
				})
			}
			// propose new commands
			for _, req := range p.requests {
				p.P2a(req)
			}
			p.requests = make([]*paxi.Request, 0)
		}
	}
}

// HandleP2a handles P2a message
func (p *Paxos) HandleP2a(m P2a) {
	log.Debugf("Replica %s ===[%v]===>>> Replica %s\n", m.Ballot.ID(), m, p.ID())

	if m.Ballot >= p.ballot {
		p.ballot = m.Ballot
		p.globalExecuted = m.GlobalExecutedSlot
		p.active = false
		// update slot number
		p.slot = paxi.Max(p.slot, m.Slot)
		// update entry
		p.logLck.Lock()
		if e, exists := p.log[m.Slot]; exists {
			if !e.commit && m.Ballot > e.ballot {
				// different command and request is not nil
				if !e.command.Equal(m.Command) && e.request != nil {
					p.Forward(m.Ballot.ID(), *e.request)
					// p.Retry(*e.request)
					log.Debugf("Received different command (%v!=%v) for slot %d, resetting the request", m.Command, e.command, m.Slot)
					e.request = nil
				}
				e.command = m.Command
				e.ballot = m.Ballot
			} else if e.commit && e.ballot == 0 {
				// we can have commit slot with no ballot when we received P3 before P2a
				e.command = m.Command
				e.ballot = m.Ballot
			}
		} else {
			p.log[m.Slot] = &entry{
				ballot:  m.Ballot,
				command: m.Command,
				commit:  false,
			}
		}
		p.logLck.Unlock()
	}

	p.Send(m.Ballot.ID(), P2b{
		Ballot: 		  p.ballot,
		Slot:   		  m.Slot,
		ID:     		  p.ID(),
		LastExecutedSlot: p.execute,
	})

	if len(m.P3msg.Slot) > 0 {
		p.HandleP3(m.P3msg)
	}
}

// HandleP2b handles P2b message
func (p *Paxos) HandleP2b(m P2b) {
	// old message

	p.logLck.RLock()
	entry, exist := p.log[m.Slot]
	p.logLck.RUnlock()
	if !exist || m.Ballot < entry.ballot || entry.commit {
		return
	}

	p.UpdateLastExecutedByNode(m.ID, m.LastExecutedSlot)
	// reject message
	// node update its ballot number and falls back to acceptor
	if m.Ballot > p.ballot {
		p.ballot = m.Ballot
		p.active = false
		// send pending P3s we have for old ballot
		p.p3Lock.Lock()
		p.Broadcast(P3{
			Ballot:  p.p3PendingBallot,
			Slot:    p.p3pendingSlots,
		})
		p.lastP3Time = hlc.CurrentTimeInMS()
		p.p3pendingSlots = make([]int, 0, 100)
		p.p3PendingBallot = 0
		p.p3Lock.Unlock()
	}

	// ack message
	// the current slot might still be committed with q2
	// if no q2 can be formed, this slot will be retried when received p2a or p3
	if m.Ballot.ID() == p.ID() && m.Ballot == entry.ballot {
		entry.quorum.ACK(m.ID)
		if p.Q2(entry.quorum) {
			entry.commit = true

			p.p3Lock.Lock()
			log.Debugf("Adding slot %d to P3Pending (%v)", m.Slot, p.p3pendingSlots)
			p.p3pendingSlots = append(p.p3pendingSlots, m.Slot)
			p.p3Lock.Unlock()

			if p.ReplyWhenCommit {
				r := entry.request
				r.Reply(paxi.Reply{
					Command:   r.Command,
					Timestamp: r.Timestamp,
				})
			} else {
				p.exec()
			}
		}
	}
}

// HandleP3 handles phase 3 commit message
func (p *Paxos) HandleP3(m P3) {
	log.Debugf("Replica %s HandleP3 {%v} from %v", p.ID(), m, m.Ballot.ID())
	for _, slot := range m.Slot {
		p.slot = paxi.Max(p.slot, slot)
		p.logLck.Lock()
		e, exist := p.log[slot]
		if exist {
			if e.ballot == m.Ballot {
				e.commit = true
			} else if e.request != nil {
				// p.Retry(*e.request)
				p.Forward(m.Ballot.ID(), *e.request)
				e.request = nil
				// ask to recover the slot
				log.Debugf("Replica %s needs to recover slot %d on ballot %v", p.ID(), slot, m.Ballot)
				recoverSlots := []int{slot}
				p.sendRecoverRequest(m.Ballot, recoverSlots)
			}

		} else {
			// we mark slot as committed, but set ballot to 0 to designate that we have not received P2a for the slot and may need to recover later
			e = &entry{commit: true, ballot: 0}
			p.log[slot] = e
		}
		p.logLck.Unlock()

		if p.ReplyWhenCommit {
			if e.request != nil {
				e.request.Reply(paxi.Reply{
					Command:   e.request.Command,
					Timestamp: e.request.Timestamp,
				})
			}
		}
	}
	p.exec()
	//log.Debugf("Leaving HandleP3")
}

// HandleP3 handles phase 3 commit message
/*func (p *Paxos) HandleP3(m P3) {
	// log.Debugf("Replica %s ===[%v]===>>> Replica %s\n", m.Ballot.ID(), m, p.ID())

	p.slot = paxi.Max(p.slot, m.Slot)

	e, exist := p.log[m.Slot]
	if exist {
		if !e.command.Equal(m.Command) && e.request != nil {
			// p.Retry(*e.request)
			p.Forward(m.Ballot.ID(), *e.request)
			e.request = nil
		}
	} else {
		p.log[m.Slot] = &entry{}
		e = p.log[m.Slot]
	}

	e.command = m.Command
	e.commit = true

	if p.ReplyWhenCommit {
		if e.request != nil {
			e.request.Reply(paxi.Reply{
				Command:   e.request.Command,
				Timestamp: e.request.Timestamp,
			})
		}
	} else {
		p.exec()
	}
}*/

func (p *Paxos) sendRecoverRequest(ballot paxi.Ballot, slots []int) {
	p.Send(ballot.ID(), P3RecoverRequest{
		Ballot: ballot,
		Slots:  slots,
		NodeId: p.ID(),
	})
}


// HandleP3RecoverRequest handles slot recovery request at leader
func (p *Paxos) HandleP3RecoverRequest(m P3RecoverRequest) {
	log.Debugf("Replica %s ===[%v]===>>> Replica %s\n", m.NodeId, m, p.ID())
	p.logLck.Lock()

	slotsToRecover := make([]int, 0, len(m.Slots))
	cmdsToRecover := make([]paxi.Command, 0, len(m.Slots))
	for _, slot := range m.Slots {
		e, exist := p.log[slot]
		if exist && e.commit {
			log.Debugf("Entry on slot %d for recover: %v", slot, e)
			slotsToRecover = append(slotsToRecover, slot)
			cmdsToRecover = append(cmdsToRecover, e.command)
		} else {
			log.Debugf("Entry for recovery on slot %d does not exist or uncommitted", slot)
		}
	}
	p.logLck.Unlock()

	// ok to recover
	log.Debugf("Node %v sends recovery on slots %d to node %v", p.ID(), slotsToRecover, m.NodeId)
	p.Send(m.NodeId, P3RecoverReply{
		Ballot:   p.ballot, // TODO: this technically needs to be a list of ballots for each slot, but I think there is no harm to make recovered slots with a higher or same ballot than the original
		Slots:    slotsToRecover,
		Commands: cmdsToRecover,
	})


	log.Debugf("Leaving HandleP3RecoverRequest")
}

// HandleP3RecoverReply handles slot recovery
func (p *Paxos) HandleP3RecoverReply(m P3RecoverReply) {
	log.Debugf("[%v]===>>> Replica %s\n",  m, p.ID())
	p.logLck.Lock()

	for i, slot := range m.Slots {
		p.slot = paxi.Max(p.slot, slot)

		// overwrite the slot with one we recovered, as it is guaranteed to have been majority committed
		p.log[slot] = &entry{
			ballot:    m.Ballot,
			command:   m.Commands[i],
			commit:    true,
			timestamp: time.Now(),
		}
	}

	p.logLck.Unlock()

	p.exec()
	log.Debugf("Leaving HandleP3RecoverReply")
}

func (p *Paxos) exec() {
	p.logLck.Lock()
	defer p.logLck.Unlock()
	for {
		e, exists := p.log[p.execute]

		if !exists || !e.commit {
			break
		}
		log.Debugf("Replica %s execute [s=%d, cmd=%v]", p.ID(), p.execute, e.command)
		value := p.Execute(e.command)
		if e.request != nil {
			reply := paxi.Reply{
				Command:    e.command,
				Value:      value,
				Properties: make(map[string]string),
			}
			reply.Properties[HTTPHeaderSlot] = strconv.Itoa(p.execute)
			reply.Properties[HTTPHeaderBallot] = e.ballot.String()
			reply.Properties[HTTPHeaderExecute] = strconv.Itoa(p.execute)
			e.request.Reply(reply)
			e.request = nil
		}
		// TODO clean up the log periodically
		// delete(p.log, p.execute)
		p.execute++
	}
	log.Debugf("Leaving exec")
}

func (p *Paxos) forward() {
	for _, m := range p.requests {
		p.Forward(p.ballot.ID(), *m)
	}
	p.requests = make([]*paxi.Request, 0)
}
