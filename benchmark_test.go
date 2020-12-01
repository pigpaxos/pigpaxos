package paxi

import (
	"sync"
	"testing"

	"github.com/pigpaxos/pigpaxos/log"
)

type FakeDB struct {
	start int
	end   int
	local int
	total int

	lock sync.Mutex
}

func (f *FakeDB) Init() error {
	return nil
}

func (f *FakeDB) Stop() error {
	log.Infof("local / total = %d / %d = %f", f.local, f.total, float64(f.local)/float64(f.total))
	return nil
}

func (f *FakeDB) Read(key int) ([]byte, error) {
	//log.Debugf("Read %d", key)
	f.lock.Lock()
	f.total++
	if key >= f.start && key <= f.end {
		f.local++

	}
	f.lock.Unlock()
	b := make([]byte, 8)
	return b, nil
}

func (f *FakeDB) Write(key int, value []byte) error {
	//log.Debugf("Write %d", key)
	f.lock.Lock()
	f.total++
	if key >= f.start && key <= f.end {
		f.local++

	}
	f.lock.Unlock()
	return nil
}

func TestBenchmark(t *testing.T) {
	start := 200
	end := 400

	f := new(FakeDB)
	f.start = start
	f.end = end

	b := NewBenchmark(f)
	b.Min = start
	b.K = 1000
	b.Distribution = "normal"
	b.Mu = 300
	b.Sigma = 50
	b.T = 0
	b.N = 10000
	b.LinearizabilityCheck = false

	b.Run()
}
