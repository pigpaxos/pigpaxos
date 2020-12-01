package main

import (
	"flag"
	"github.com/pigpaxos/pigpaxos/epaxos"
	"github.com/pigpaxos/pigpaxos"
	"github.com/pigpaxos/pigpaxos/paxos"
)

var id = flag.String("id", "", "node id this client connects to")
var algorithm = flag.String("algorithm", "", "Client API type [paxos, chain]")
var load = flag.Bool("load", false, "Load K keys into DB")
var master = flag.String("master", "", "Master address.")

// db implements Paxi.DB interface for benchmarking
type db struct {
	paxi.Client
}

func (d *db) Init() error {
	return nil
}

func (d *db) Stop() error {
	return nil
}

func (d *db) Read(k int) ([]byte, error) {
	key := paxi.Key(k)
	v, err := d.Get(key)
	if len(v) == 0 {
		return nil, nil
	}
	return v, err
}

func (d *db) Write(k int, v []byte) error {
	key := paxi.Key(k)
	err := d.Put(key, v)
	return err
}

func main() {
	paxi.Init()

	if *master != "" {
		paxi.ConnectToMaster(*master, true, paxi.NewIDFromString(*id))
	}

	d := new(db)
	switch *algorithm {
	case "paxos":
		d.Client = paxos.NewClient(paxi.NewIDFromString(*id))
	case "epaxos":
		d.Client = epaxos.NewClient(paxi.NewIDFromString(*id))
	default:
		d.Client = paxi.NewHTTPClient(paxi.NewIDFromString(*id))
	}

	b := paxi.NewBenchmark(d)
	if *load {
		b.Load()
	} else {
		b.Run()
	}
}
