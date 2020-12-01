package paxi

import (
	"bytes"
	"encoding/gob"
	"strconv"
	"testing"
)

type A struct {
	I int
	S string
	B bool
}

type B struct {
	S string
}

type testP2b struct {
	ID      []ID // from node id
	Ballot  Ballot
	Slot    int
}

type testP2bInt struct {
	ID      []uint32 // from node id
	Ballot  Ballot
	Slot    int
}

type testP2bAggregate struct {
	MissingIDs       []ID // node ids not collected by relay
	RelayID			 ID
	RelayLastExecute int
	Ballot  		 Ballot
	Slot    		 int
}

type testP2a struct {
	Ballot  Ballot
	Slot    int
	Command Command
}

type PreAccept struct {
	Ballot  Ballot
	Replica ID
	Slot    int
	Command Command
	Seq     int
	Dep     map[ID]int
}

func TestCodecGob(t *testing.T) {
	gob.Register(A{})
	gob.Register(B{})
	var send interface{}
	var recv interface{}

	buf := new(bytes.Buffer)
	c := NewCodec("gob", buf)

	send = A{1, "a", true}

	c.Encode(&send)
	c.Decode(&recv)
	if send.(A) != recv.(A) {
		t.Errorf("expect send %v and recv %v to be euqal", send, recv)
	}

	send = B{"test"}

	c.Encode(&send)
	c.Decode(&recv)
	if send.(B) != recv.(B) {
		t.Errorf("expect send %v and recv %v to be euqal", send, recv)
	}
}

func BenchmarkCodecGob(b *testing.B) {
	gob.Register(A{})
	var send interface{}
	var recv interface{}

	buf := new(bytes.Buffer)
	c := NewCodec("gob", buf)

	send = A{1, "a", true}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.Encode(&send)
		c.Decode(&recv)
	}
}

func BenchmarkCodecGobP2bInt(b *testing.B) {
	gob.Register(testP2bInt{})
	var send interface{}
	var recv interface{}
	buf := new(bytes.Buffer)
	c := NewCodec("gob", buf)

	ids := make([]uint32, 3)
	ids = append(ids, 1)
	ids = append(ids, 2)
	ids = append(ids, 3)
	ids = append(ids, 4)
	ids = append(ids, 5)
	ids = append(ids, 6)
	ids = append(ids, 7)
	/*ids = append(ids, "1.8")
	ids = append(ids, "1.9")
	ids = append(ids, "1.10")
	ids = append(ids, "1.11")*/

	bal := NewBallot(0, NewIDFromString("1.1"))

	send = testP2bInt{ids, bal, 0}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.Encode(&send)
		c.Decode(&recv)
	}
}

func BenchmarkCodecGobP2b(b *testing.B) {
	gob.Register(testP2b{})
	var send interface{}
	var recv interface{}
	buf := new(bytes.Buffer)
	c := NewCodec("gob", buf)

	ids := make([]ID, 3)
	ids = append(ids, 1)
	ids = append(ids, 2)
	ids = append(ids, 3)
	ids = append(ids, 4)
	ids = append(ids, 5)
	ids = append(ids, 6)
	ids = append(ids, 7)
	/*ids = append(ids, "1.8")
	ids = append(ids, "1.9")
	ids = append(ids, "1.10")
	ids = append(ids, "1.11")*/

	bal := NewBallot(0, NewIDFromString("1.1"))

	send = testP2b{ids, bal, 0}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.Encode(&send)
		c.Decode(&recv)
	}
}

func BenchmarkCodecGobSerializeP2b(b *testing.B) {
	gob.Register(testP2b{})
	var send interface{}

	buf := new(bytes.Buffer)
	c := NewCodec("gob", buf)

	ids := make([]ID, 3)
	ids = append(ids, NewIDFromString("1.1"))
	ids = append(ids, NewIDFromString("1.2"))
	ids = append(ids, NewIDFromString("1.3"))
	ids = append(ids, NewIDFromString("1.4"))
	ids = append(ids, NewIDFromString("1.5"))
	ids = append(ids, NewIDFromString("1.6"))
	ids = append(ids, NewIDFromString("1.7"))
	/*ids = append(ids, "1.8")
	ids = append(ids, "1.9")
	ids = append(ids, "1.10")
	ids = append(ids, "1.11")*/

	bal := NewBallot(0, NewIDFromString("1.1"))

	send = testP2b{ids, bal, 0}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.Encode(&send)
	}
}

func BenchmarkCodecGobSerializeP2bAggregate(b *testing.B) {
	gob.Register(testP2bAggregate{})
	var send interface{}

	buf := new(bytes.Buffer)
	c := NewCodec("gob", buf)

	missingIds := make([]ID, 0)

	bal := NewBallot(0, NewIDFromString("1.1"))

	send = testP2bAggregate{RelayLastExecute: 42, RelayID: NewIDFromString("1.1"), Ballot: bal, Slot: 42, MissingIDs: missingIds}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.Encode(&send)
	}
}

func BenchmarkCodecGobP2bAggregate(b *testing.B) {
	gob.Register(testP2bAggregate{})
	var send interface{}
	var recv interface{}

	buf := new(bytes.Buffer)
	c := NewCodec("gob", buf)

	missingIds := make([]ID, 0)

	bal := NewBallot(0, NewIDFromString("1.1"))

	send = testP2bAggregate{RelayLastExecute: 42, RelayID: NewIDFromString("1.1"), Ballot: bal, Slot: 42, MissingIDs: missingIds}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.Encode(&send)
		c.Decode(&recv)
	}
}

func BenchmarkCodecGobSerializeP2a(b *testing.B) {
	gob.Register(testP2a{})
	var send interface{}

	buf := new(bytes.Buffer)
	c := NewCodec("gob", buf)

	id := NewIDFromString("1.1")
	bal := NewBallot(1, id)

	v1 := GenerateRandVal(1)

	Cmd := Command{ClientID: NewIDFromString("1.1"), CommandID: 1, Key: 1, Value: v1}

	send = testP2a{bal, 1, Cmd }

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.Encode(&send)
	}
}

func BenchmarkCodecGobSerializePreAccept(b *testing.B) {
	gob.Register(PreAccept{})
	var send interface{}
	var recv interface{}

	buf := new(bytes.Buffer)
	c := NewCodec("gob", buf)


	bal := NewBallot(0, NewIDFromString("1.1"))
	v1 := GenerateRandVal(1)

	Cmd := Command{ClientID: NewIDFromString("1.1"), CommandID: 1, Key: 1, Value: v1}

	deps := make(map[ID]int, 0)
	for i := 1; i <= 25; i++ {
		id := NewIDFromString("1." + strconv.Itoa(i))
		deps[id] = 100 + i
	}

	send = PreAccept{bal, NewIDFromString("1.1"), 0, Cmd, 0, deps}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.Encode(&send)
		c.Decode(&recv)
	}
}

func BenchmarkCodecJSON(b *testing.B) {
	buf := new(bytes.Buffer)
	var send interface{}
	var recv interface{}

	c := NewCodec("json", buf)

	send = A{1, "a", true}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.Encode(&send)
		c.Decode(&recv)
	}
}
