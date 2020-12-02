package pigpaxos

import (
	"github.com/pigpaxos/pigpaxos/hlc"
	"github.com/pigpaxos/pigpaxos/retro_log"
	"time"

	"github.com/pigpaxos/pigpaxos"
	"github.com/pigpaxos/pigpaxos/log"
	"sync"
)

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
type PigPaxos struct {
	paxi.Node

	// log management variables
	log           		map[int]*entry  // log ordered by slot
	slot          		int             // highest slot number
	execute       		int             // next execute slot number
	lastCleanupMarker 	int
	globalExecute 		int             // executed by all nodes. Need for log cleanup
	executeByNode 		map[paxi.ID]int // leader's knowledge of other nodes execute counter. Need for log cleanup

	//Paxos management
	active        	bool            // active leader
	ballot        	paxi.Ballot     // highest ballot number
	p1aTime       	int64           // time P1a was sent
	quorum   		*paxi.Quorum    // phase 1 quorum
	requests 		[]*paxi.Request // phase 1 pending requests

	p3PendingBallot paxi.Ballot
	p3pendingSlots  []int
	lastP3Time      int64


	// Quorums
	Q1              func(*paxi.Quorum) bool
	Q2              func(*paxi.Quorum) bool
	ReplyWhenCommit bool

	// Locks
	logLck			sync.RWMutex
	p3Lock			sync.RWMutex
	markerLock		sync.RWMutex
}

// NewPaxos creates new paxos instance
func NewPigPaxos(n paxi.Node, options ...func(*PigPaxos)) *PigPaxos {
	p := &PigPaxos{
		Node:            n,
		log:             make(map[int]*entry, paxi.GetConfig().BufferSize),
		slot:            -1,
		quorum:          paxi.NewQuorum(),
		requests:        make([]*paxi.Request, 0),
		p3pendingSlots:  make([]int,0, 100),
		executeByNode:   make(map[paxi.ID]int, 0),
		lastP3Time:      0,
		Q1:              func(q *paxi.Quorum) bool { return q.Majority() },
		Q2:              func(q *paxi.Quorum) bool { return q.Majority() },
		ReplyWhenCommit: false,
	}

	for _, opt := range options {
		opt(p)
	}

	return p
}

// IsLeader indicates if this node is current leader
func (p *PigPaxos) IsLeader() bool {
	return p.active || p.ballot.ID() == p.ID()
}

func (p *PigPaxos) CheckTimeout(timeout int64) {
	p.logLck.RLock()
	defer p.logLck.RUnlock()
	execslot := p.execute
	if p.active && execslot <= p.slot {
		if e, ok := p.log[execslot]; ok && !e.commit {
			if e.timestamp.UnixNano() < timeout {
				log.Debugf("Retrying p2. entry_time = %d, retry time = %d", e.timestamp, timeout)
				p.RetryP2a(execslot, e)
			}
		}
	} else if !p.active && p.p1aTime < timeout {
		log.Debugf("Retrying p1. p1time = %d, retry time = %d", p.p1aTime, timeout)
		p.RetryP1a()
	}

	tnow := hlc.CurrentTimeInMS()
	if tnow - 10 > p.lastP3Time && p.lastP3Time > 0 && p.p3PendingBallot > 0 {
		log.Debugf("Sending P3 on timeout: %v", p.p3PendingBallot)
		p.p3Lock.Lock()
		p.Broadcast(P3{
			Ballot:  p.p3PendingBallot,
			Slot:    p.p3pendingSlots,
		})
		p.lastP3Time = hlc.CurrentTimeInMS()
		p.p3pendingSlots = make([]int,0, 100)
		p.p3Lock.Unlock()
	}
}

// Leader returns leader id of the current ballot
func (p *PigPaxos) Leader() paxi.ID {
	return p.ballot.ID()
}

// Ballot returns current ballot
func (p *PigPaxos) Ballot() paxi.Ballot {
	return p.ballot
}

// SetActive sets current paxos instance as active leader
func (p *PigPaxos) SetActive(active bool) {
	p.active = active
}

// SetBallot sets a new ballot number
func (p *PigPaxos) SetBallot(b paxi.Ballot) {
	p.ballot = b
}

func (p *PigPaxos) UpdateLastExecuteByNode(id paxi.ID, lastExecute int) {
	p.markerLock.Lock()
	defer p.markerLock.Unlock()
	p.executeByNode[id] = lastExecute
}

func (p *PigPaxos) GetSafeLogCleanupMarker() int {
	marker := p.execute
	for _, c := range p.executeByNode {
		if c < marker {
			marker = c
		}
	}
	return marker
}

func (p *PigPaxos) CleanupLog() {
	p.markerLock.Lock()
	marker := p.GetSafeLogCleanupMarker()
	//log.Debugf("Replica %v log cleanup. lastCleanupMarker: %d, safeCleanUpMarker: %d", p.ID(), p.lastCleanupMarker, marker)
	p.markerLock.Unlock()

	p.logLck.Lock()
	defer p.logLck.Unlock()
	for i := p.lastCleanupMarker; i < marker; i++ {
		delete(p.log, i)
	}
	p.lastCleanupMarker = marker
}

// HandleRequest handles request and start phase 1 or phase 2
func (p *PigPaxos) HandleRequest(r paxi.Request) {
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
func (p *PigPaxos) P1a() {
	log.Debugf("Node %v PigPaxos P1a", p.ID())
	if p.active {
		return
	}
	p.ballot.Next(p.ID())
	p.quorum.Reset()
	p.quorum.ACK(p.ID())
	p.p1aTime = time.Now().UnixNano()
	p.Broadcast(P1a{Ballot: p.ballot})
}

func (p *PigPaxos) RetryP1a() {
	log.Debugf("Node %v PigPaxos RetryP1a", p.ID())
	// should be called from within a lock
	if p.active {
		return
	}
	p.quorum.Reset()
	p.quorum.ACK(p.ID())
	p.p1aTime = time.Now().UnixNano()
	p.Broadcast(P1a{Ballot: p.ballot})
}

// P2a starts phase 2 accept
func (p *PigPaxos) P2a(r *paxi.Request) {
	log.Debugf("Node %v etering P2a with slot %d", p.ID(), p.slot)
	p.logLck.Lock()
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
		Ballot:  p.ballot,
		Slot:    p.slot,
		Command: r.Command,
		GlobalExecute: p.globalExecute,
	}
	p.logLck.Unlock()
	p.p3Lock.Lock()
	if p.p3PendingBallot > 0 {
		m.P3msg = P3{Ballot: p.p3PendingBallot, Slot: p.p3pendingSlots}
		p.p3pendingSlots = make([]int,0, 100)
		p.lastP3Time = hlc.CurrentTimeInMS()
	}
	p.p3Lock.Unlock()

	if paxi.GetConfig().Thrifty {
		go p.Broadcast(m) // TODO: implement thrifty
	} else {
		go p.Broadcast(m)
	}
	log.Debugf("Leaving P2a with slot %d", p.slot)
}

func (p *PigPaxos) RetryP2a(slot int, e *entry) {
	log.Debugf("Entering RetryP2a with slot %d", slot)
	m := P2a{
		Ballot:  p.ballot,
		Slot:    slot,
		Command: e.command,
		GlobalExecute: p.globalExecute,
	}
	if paxi.GetConfig().Thrifty {
		p.Broadcast(m) // TODO: implement thrifty
	} else {
		p.Broadcast(m)
	}
	log.Debugf("Leaving RetryP2a with slot %d", p.slot)
}

// HandleP1a handles P1a message
func (p *PigPaxos) HandleP1a(m P1a, reply paxi.ID) {
	log.Debugf("Replica %s ===[%v]===>>> Replica %s\n", m.Ballot.ID(), m, p.ID())

	// new leader
	if m.Ballot > p.ballot {
		p.ballot = m.Ballot
		p.active = false
		p.forward()
	}

	l := make(map[int]CommandBallot)
	p.logLck.RLock()
	for s := p.execute; s <= p.slot; s++ {
		if p.log[s] == nil || p.log[s].commit {
			continue
		}
		l[s] = CommandBallot{p.log[s].command, p.log[s].ballot}
	}
	p.logLck.RUnlock()

	p.Send(reply, P1b{
		Ballot: p.ballot,
		ID:     p.ID(),
		Log:    l,
	})
	log.Debugf("Leaving HandleP1a")
}

func (p *PigPaxos) update(scb map[int]CommandBallot) {
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
func (p *PigPaxos) HandleP1b(m P1b) {
	// old message
	if m.Ballot < p.ballot || p.active {
		return
	}

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
			p.logLck.Lock()
			for i := p.execute; i <= p.slot; i++ {
				// TODO nil gap?
				if p.log[i] == nil || p.log[i].commit {
					continue
				}
				p.log[i].ballot = p.ballot
				p.log[i].quorum = paxi.NewQuorum()
				p.log[i].quorum.ACK(p.ID())
				p.Broadcast(P2a{
					Ballot:  p.ballot,
					Slot:    i,
					Command: p.log[i].command,
					GlobalExecute: p.globalExecute,
				})
			}
			p.logLck.Unlock()
			// propose new commands
			for _, req := range p.requests {
				p.P2a(req)
			}
			p.requests = make([]*paxi.Request, 0)
		}
	}
}

// HandleP2a handles P2a message
func (p *PigPaxos) HandleP2a(m P2a, reply paxi.ID) {
	log.Debugf("Replica %s ===[%v]===>>> Replica %s\n", m.Ballot.ID(), m, p.ID())

	if m.Ballot >= p.ballot {
		p.ballot = m.Ballot
		p.active = false
		p.globalExecute = m.GlobalExecute
		p.logLck.Lock()
		// update slot number
		p.slot = paxi.Max(p.slot, m.Slot)
		// update entry
		if e, exists := p.log[m.Slot]; exists {
			if !e.commit && m.Ballot > e.ballot {
				// different command and request is not nil
				if !e.command.Equal(m.Command) && e.request != nil {
					p.Forward(m.Ballot.ID(), *e.request)
					// p.Retry(*e.request)
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

	idList := make([]paxi.ID, 1, 1)
	idList[0] = p.ID()

	p.Send(reply, P2b{
		Ballot: p.ballot,
		Slot:   m.Slot,
		ID:     idList,
	})

	if len(m.P3msg.Slot) > 0 {
		p.HandleP3(m.P3msg)
	}

	log.Debugf("Leaving HandleP2a")
}

// HandleP2b handles P2b message
func (p *PigPaxos) HandleP2b(msgSlot int, msgBallot paxi.Ballot, votedIds []paxi.ID) {
	log.Debugf("Entering HandleP2b: ===[bal: %v, slot: %d, votes: %v]===>>> %s", msgBallot, msgSlot, votedIds, p.ID())
	// old message
	p.logLck.RLock()
	entry, exist := p.log[msgSlot]
	p.logLck.RUnlock()
	if !exist || msgBallot < entry.ballot || entry.commit {
		return
	}
	// reject message
	// node update its ballot number and falls back to acceptor
	if msgBallot > p.ballot {
		p.ballot = msgBallot
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
	if msgBallot.ID() == p.ID() && msgBallot == entry.ballot {
		for _, id := range votedIds {
			entry.quorum.ACK(id)
		}

		if p.Q2(entry.quorum) {
			entry.commit = true
			if paxi.GetConfig().UseRetroLog {
				slotStruct := retro_log.NewRqlStruct(nil).AddVarInt32("slot", msgSlot).AddVarStr("hash", entry.command.Hash())
				paxi.Retrolog.StartTx().AppendSetStruct("committed", slotStruct).AppendSetInt32("committed_slots", msgSlot).Commit()
			}

			p.p3Lock.Lock()
			log.Debugf("Adding slot %d to P3Pending (%v)", msgSlot, p.p3pendingSlots)
			p.p3pendingSlots = append(p.p3pendingSlots, msgSlot)
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
	log.Debugf("Leaving HandleP2b. next execute slot = %d, next slot=%d", p.execute, p.slot)
}

// HandleP3 handles phase 3 commit message
func (p *PigPaxos) HandleP3(m P3) {
	log.Debugf("Replica %s ===[%v]===>>> Replica %s\n", m.Ballot.ID(), m, p.ID())
	for _, slot := range m.Slot {
		p.logLck.Lock()
		p.slot = paxi.Max(p.slot, slot)
		e, exist := p.log[slot]
		if exist {
			if e.ballot == m.Ballot {
				e.commit = true
			} else if e.request != nil {
				// p.Retry(*e.request)
				p.Forward(m.Ballot.ID(), *e.request)
				e.request = nil
				// ask to recover the slot
				log.Debugf("Replica %s needs to recover slot %d on ballot %v (we have cmd %v)", p.ID(), slot, m.Ballot, e.command)
				p.sendRecoverRequest(m.Ballot, slot)
			}

		} else {
			e = &entry{commit: true, ballot: 0}
			p.log[slot] = e
		}
		p.logLck.Unlock()

		if paxi.GetConfig().UseRetroLog {
			slotStruct := retro_log.NewRqlStruct(nil).AddVarInt32("slot", slot).AddVarStr("hash", e.command.Hash())
			paxi.Retrolog.StartTx().AppendSetStruct("committed", slotStruct).AppendSetInt32("committed_slots", slot).Commit()
		}
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
	log.Debugf("Leaving HandleP3")
}

func (p *PigPaxos) sendRecoverRequest(ballot paxi.Ballot, slot int) {
	p.Send(ballot.ID(), P3RecoverRequest{
		Ballot: ballot,
		Slot:   slot,
		NodeId: p.ID(),
	})
}

// HandleP3RecoverRequest handles slot recovery request at leader
func (p *PigPaxos) HandleP3RecoverRequest(m P3RecoverRequest) {
	log.Debugf("Replica %s ===[%v]===>>> Replica %s\n", m.NodeId, m, p.ID())
	p.logLck.Lock()
	e, exist := p.log[m.Slot]
	p.logLck.Unlock()
	if exist && e.commit {
		// ok to recover
		p.Send(m.NodeId, P3RecoverReply{
			Ballot: e.ballot,
			Slot:   m.Slot,
			Command:e.command,
		})
	}

	p.exec()
	log.Debugf("Leaving HandleP3RecoverRequest")
}

// HandleP3RecoverReply handles slot recovery
func (p *PigPaxos) HandleP3RecoverReply(m P3RecoverReply) {
	log.Debugf("Replica %s ===[%v]===>>> Replica %s\n", m.Ballot.ID(), m, p.ID())
	p.logLck.Lock()
	p.slot = paxi.Max(p.slot, m.Slot)
	e, exist := p.log[m.Slot]
	if exist {
		e.command = m.Command
		e.ballot = m.Ballot
		e.commit = true
	}
	p.logLck.Unlock()

	p.exec()
	log.Debugf("Leaving HandleP3RecoverReply")
}


func (p *PigPaxos) exec() {
	log.Debugf("Entering exec. exec slot=%d", p.execute)
	p.logLck.Lock()
	defer p.logLck.Unlock()
	for {
		e, ok := p.log[p.execute]
		if ok && p.execute + 10 < p.slot && e.commit && e.ballot == 0 {
			// ask to recover the slot
			log.Debugf("Replica %s tries to recover slot %d on ballot %v", p.ID(),  p.execute, p.Ballot())
			p.sendRecoverRequest(p.Ballot(), p.execute)
		}

		if !ok || !e.commit || (e.commit && e.ballot == 0) {
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
			go e.request.Reply(reply)
			e.request = nil
		}
		p.execute++
	}
	log.Debugf("Leaving exec")
}

func (p *PigPaxos) forward() {
	for _, m := range p.requests {
		p.Forward(p.ballot.ID(), *m)
	}
	p.requests = make([]*paxi.Request, 0)
}
