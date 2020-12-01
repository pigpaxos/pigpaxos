package paxos

import (
	"encoding/gob"
	"fmt"

	"github.com/pigpaxos/pigpaxos"
)

func init() {
	gob.Register(P1a{})
	gob.Register(P1b{})
	gob.Register(P2a{})
	gob.Register(P2b{})
	gob.Register(P3{})
}

// P1a prepare message
type P1a struct {
	Ballot paxi.Ballot
}

func (m P1a) String() string {
	return fmt.Sprintf("P1a {b=%v}", m.Ballot)
}

// CommandBallot conbines each command with its ballot number
type CommandBallot struct {
	Command paxi.Command
	Ballot  paxi.Ballot
}

func (cb CommandBallot) String() string {
	return fmt.Sprintf("cmd=%v b=%v", cb.Command, cb.Ballot)
}

// P1b promise message
type P1b struct {
	Ballot paxi.Ballot
	ID     paxi.ID               // from node id
	Log    map[int]CommandBallot // uncommitted logs
}

func (m P1b) String() string {
	return fmt.Sprintf("P1b {b=%v id=%s log=%v}", m.Ballot, m.ID, m.Log)
}

// P2a accept message
type P2a struct {
	Ballot  			 paxi.Ballot
	Slot    			 int
	GlobalExecutedSlot   int
	Command 			 paxi.Command
	P3msg				 P3
}

func (m P2a) String() string {
	return fmt.Sprintf("P2a {b=%v s=%d cmd=%v, P3={%v}}", m.Ballot, m.Slot, m.Command, m.P3msg)
}

// P2b accepted message
type P2b struct {
	Ballot paxi.Ballot
	ID     paxi.ID // from node id
	Slot   int
	LastExecutedSlot int
}

func (m P2b) String() string {
	return fmt.Sprintf("P2b {b=%v id=%s s=%d, lastexec=%d}", m.Ballot, m.ID, m.Slot, m.LastExecutedSlot)
}

// P3 commit message
type P3 struct {
	Ballot  paxi.Ballot
	Slot    []int
	Command paxi.Command
}

func (m P3) String() string {
	return fmt.Sprintf("P3 {b=%v s=%d cmd=%v}", m.Ballot, m.Slot, m.Command)
}


type P3RecoverRequest struct {
	Ballot paxi.Ballot
	Slots  []int
	NodeId paxi.ID
}

func (m P3RecoverRequest) String() string {
	return fmt.Sprintf("P3RecoverRequest {b=%v Slots=%v, nodeToRecover=%v}",  m.Ballot, m.Slots, m.NodeId)
}

type P3RecoverReply struct {
	Ballot    paxi.Ballot
	Slots     []int
	Commands  []paxi.Command
}

func (m P3RecoverReply) String() string {
	return fmt.Sprintf("P3RecoverReply {b=%v slots=%v, cmd=%v}",  m.Ballot, m.Slots, m.Commands)
}