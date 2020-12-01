package pigpaxos

import (
	"github.com/pigpaxos/pigpaxos"
	"testing"
)

var replica *Replica

func testSetup() *Replica {
	replica := NewReplica("1.2")
	return replica
}

func BenchmarkMissingIDAllMissing(b *testing.B) {
	paxi.GetConfig().Addrs["1.1"] = "127.0.0.1:1735"
	paxi.GetConfig().Addrs["1.2"] = "127.0.0.1:1736"
	paxi.GetConfig().Addrs["1.3"] = "127.0.0.1:1737"
	paxi.GetConfig().Addrs["1.4"] = "127.0.0.1:1738"
	paxi.GetConfig().Addrs["1.5"] = "127.0.0.1:1739"
	paxi.GetConfig().Addrs["1.6"] = "127.0.0.1:1740"
	paxi.GetConfig().Addrs["1.7"] = "127.0.0.1:1741"
	paxi.GetConfig().Addrs["1.8"] = "127.0.0.1:1742"
	if replica == nil {
		replica = testSetup()
	}
	bal := paxi.NewBallot(0, "1.1")
	p2b := P2b{ID: make([]paxi.ID, 0), Ballot: bal, Slot: 42}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		replica.computeMissingIDsForP2b(p2b)
	}
}

func BenchmarkMissingIDOneMissing(b *testing.B) {
	paxi.GetConfig().Addrs["1.1"] = "127.0.0.1:1735"
	paxi.GetConfig().Addrs["1.2"] = "127.0.0.1:1736"
	paxi.GetConfig().Addrs["1.3"] = "127.0.0.1:1737"
	paxi.GetConfig().Addrs["1.4"] = "127.0.0.1:1738"
	paxi.GetConfig().Addrs["1.5"] = "127.0.0.1:1739"
	paxi.GetConfig().Addrs["1.6"] = "127.0.0.1:1740"
	paxi.GetConfig().Addrs["1.7"] = "127.0.0.1:1741"
	paxi.GetConfig().Addrs["1.8"] = "127.0.0.1:1742"
	if replica == nil {
		replica = testSetup()
	}
	bal := paxi.NewBallot(0, "1.1")
	pgIds := []paxi.ID{paxi.ID("1.1"), paxi.ID("1.2"), paxi.ID("1.3")}
	p2b := P2b{ID: pgIds, Ballot: bal, Slot: 42}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		replica.computeMissingIDsForP2b(p2b)
	}
}
