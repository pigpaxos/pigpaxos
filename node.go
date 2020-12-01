package paxi

import (
	"github.com/pigpaxos/pigpaxos/hlc"
	"github.com/pigpaxos/pigpaxos/retro_log"
	"net/http"
	"reflect"
	"sync"

	"github.com/pigpaxos/pigpaxos/log"
)

var Retrolog *retro_log.RetroLog

// Node is the primary access point for every replica
// it includes networking, state machine and RESTful API server
type Node interface {
	Socket
	Database
	ID() ID
	Run()
	Retry(r Request)
	Forward(id ID, r Request)
	Register(m interface{}, f interface{})
	HandleMsg(m interface{})
}

// node implements Node interface
type node struct {
	id ID

	Socket
	Database
	MessageChan chan interface{}
	handles     map[string]reflect.Value
	server      *http.Server

	recvCount	int

	sync.RWMutex
	forwards map[string]*Request
}

// NewNode creates a new Node object from configuration
func NewNode(id ID) Node {
	if config.UseRetroLog {
		Retrolog = retro_log.NewRetroLog("paxi", int(id), "logs/", 100, false)
		Retrolog.CreateTimerSet("sentM", 5)
		Retrolog.CreateTimerSet("recvM", 5)
	}
	return &node{
		id:          id,
		Socket:      NewSocket(id, config.Addrs),
		Database:    NewDatabase(),
		MessageChan: make(chan interface{}, config.ChanBufferSize),
		handles:     make(map[string]reflect.Value),
		forwards:    make(map[string]*Request),
		recvCount:   0,
	}
}

func (n *node) ID() ID {
	return n.id
}

func (n *node) Retry(r Request) {
	log.Debugf("node %v retry request %v", n.id, r)
	n.MessageChan <- r
}

// Register a handle function for each message type
func (n *node) Register(m interface{}, f interface{}) {
	t := reflect.TypeOf(m)
	fn := reflect.ValueOf(f)
	if fn.Kind() != reflect.Func || fn.Type().NumIn() != 1 || fn.Type().In(0) != t {
		panic("register handle function error")
	}
	n.handles[t.String()] = fn
}

// Run start and run the node
func (n *node) Run() {
	log.Infof("node %v start running", n.id)
	if len(n.handles) > 0 {
		go n.handle()
		go n.recv()
	}
	n.http()
}

// recv receives messages from socket and pass to messa ge channel
func (n *node) recv() {
	for {
		m := n.Recv()
		switch m := m.(type) {
		case Request:
			m.c = make(chan Reply, 1)
			go func(r Request) {
				n.Send(r.NodeID, <-r.c)
			}(m)
			n.MessageChan <- m
			continue

		case Reply:
			n.RLock()
			r := n.forwards[m.Command.String()]
			log.Debugf("node %v received reply %v (m.Command.String=%s)", n.id, m, m.Command.String())
			n.RUnlock()
			r.Reply(m)
			continue
		case ProtocolMsg:
			log.Debugf("node %v received ProtocolMsg %v", n.id, m)
			hlc.HLClock.Update(*hlc.NewTimestampI64(m.HlcTime))
			Retrolog.StartTx().AppendSetInt("recvM", m.MsgId)
			n.Lock()
			n.recvCount++
			Retrolog.AppendVarInt32("recvCount", n.recvCount).Commit()
			n.Unlock()
			n.MessageChan <- m.Msg
			continue
		}
		n.MessageChan <- m
	}
}

// handle receives messages from message channel and calls handle function using refection
func (n *node) handle() {
	for {
		msg := <-n.MessageChan
		n.HandleMsg(msg)
	}
}

func (n *node) HandleMsg(msg interface{}) {
	v := reflect.ValueOf(msg)
	name := v.Type().String()
	f, exists := n.handles[name]
	if !exists {
		log.Fatalf("no registered handle function for message type %v", name)
	}
	f.Call([]reflect.Value{v})
}

/*
func (n *node) Forward(id ID, m Request) {
	key := m.Command.Key
	url := config.HTTPAddrs[id] + "/" + strconv.Itoa(int(key))

	log.Debugf("Node %v forwarding %v to %s", n.ID(), m, id)

	method := http.MethodGet
	var body io.Reader
	if !m.Command.IsRead() {
		method = http.MethodPut
		body = bytes.NewBuffer(m.Command.Value)
	}
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		log.Error(err)
		return
	}
	req.Header.Set(HTTPClientID, string(n.id))
	req.Header.Set(HTTPCommandID, strconv.Itoa(m.Command.CommandID))
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Error(err)
		m.Reply(Reply{
			Command: m.Command,
			Err:     err,
		})
		return
	}
	defer res.Body.Close()
	if res.StatusCode == http.StatusOK {
		b, err := ioutil.ReadAll(res.Body)
		if err != nil {
			log.Error(err)
		}
		m.Reply(Reply{
			Command: m.Command,
			Value:   Value(b),
		})
	} else {
		m.Reply(Reply{
			Command: m.Command,
			Err:     errors.New(res.Status),
		})
	}
}
*/

func (n *node) Forward(id ID, m Request) {
	log.Debugf("Node %v forwarding %v to %s", n.ID(), m, id)
	m.NodeID = n.id
	n.Lock()
	n.forwards[m.Command.String()] = &m
	n.Unlock()
	n.Send(id, m)
}
