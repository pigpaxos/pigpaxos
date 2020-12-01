package main

import (
	"flag"
	"sync"

	"github.com/pigpaxos/pigpaxos"
	"github.com/pigpaxos/pigpaxos/epaxos"
	"github.com/pigpaxos/pigpaxos/log"
	"github.com/pigpaxos/pigpaxos/paxos"
	"github.com/pigpaxos/pigpaxos/pigpaxos"
)

var algorithm = flag.String("algorithm", "paxos", "Distributed algorithm")
var id = flag.String("id", "", "ID in format of Zone.Node.")
var simulation = flag.Bool("sim", false, "simulation mode")

var master = flag.String("master", "", "Master address.")

func replica(id paxi.ID) {
	if *master != "" {
		paxi.ConnectToMaster(*master, false, id)
	}

	log.Infof("node %v starting with algorithm %s", id, *algorithm)

	switch *algorithm {

	case "paxos":
		paxos.NewReplica(id).Run()

	case "epaxos":
		epaxos.NewReplica(id).Run()

	case "pigpaxos":
		pigpaxos.NewReplica(id).Run()

	default:
		panic("Unknown algorithm")
	}
}

func main() {
	paxi.Init()

	if *simulation {
		var wg sync.WaitGroup
		wg.Add(1)
		paxi.Simulation()
		for id := range paxi.GetConfig().Addrs {
			n := id
			go replica(n)
		}
		wg.Wait()
	} else {
		replica(paxi.NewIDFromString(*id))
	}
}
