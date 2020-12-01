package paxi

import (
	"encoding/binary"
	"testing"
)

// examples from https://pdos.csail.mit.edu/6.824/papers/fb-consistency.pdf
func TestLinerizabilityChecker(t *testing.T) {
	c := newChecker()

	// single operation is liearizable
	bs := make([]byte, 4)
	binary.LittleEndian.PutUint32(bs, 42)

	ops := []*operation{
		&operation{bs, nil, 0, 24},
	}
	if len(c.linearizable(ops)) != 0 {
		t.Error("expected operations to be linearizable")
	}

	// concurrent operation is linearizable
	// +--w---+
	//   +---r--+

	ops = []*operation{
		&operation{bs, nil, 0, 5},
		&operation{nil, bs, 3, 10},
	}
	if len(c.linearizable(ops)) != 0 {
		t.Error("expected operations to be linearizable")
	}

	// no dependency in graph is linearizable
	n1 := make([]byte, 4)
	binary.LittleEndian.PutUint32(n1, 1)

	n2 := make([]byte, 4)
	binary.LittleEndian.PutUint32(n2, 2)

	n3 := make([]byte, 4)
	binary.LittleEndian.PutUint32(n3, 3)

	n4 := make([]byte, 4)
	binary.LittleEndian.PutUint32(n4, 4)

	ops = []*operation{
		&operation{n1, nil, 0, 5},
		&operation{nil, n2, 6, 10},
		&operation{n3, nil, 11, 15},
		&operation{nil, n4, 16, 20},
	}
	if len(c.linearizable(ops)) != 0 {
		t.Error("expected operations to be linearizable")
	}

	// concurrent reads are linearizable
	// +-------w100---------+
	//    +--r100--+
	//       +----r0-----+
	n0 := make([]byte, 4)
	binary.LittleEndian.PutUint32(n0, 0)
	n100 := make([]byte, 4)
	binary.LittleEndian.PutUint32(n100, 100)

	ops = []*operation{
		&operation{n0, nil, 0, 0},
		&operation{n100, nil, 0, 100},
		&operation{nil, n100, 5, 35},
		&operation{nil, n0, 30, 60},
	}
	if len(c.linearizable(ops)) != 0 {
		t.Error("expected operations to be linearizable")
	}

	// non-concurrent reads are not linearizable
	// +---------w100-----------+
	//   +---r100---+  +-r0--+
	ops = []*operation{
		&operation{n0, nil, 0, 0},
		&operation{n100, nil, 0, 100},
		&operation{nil, n100, 5, 25}, // reads 100, write time is cut to 25
		&operation{nil, n0, 30, 60},  // happens after previous read, but read 0
	}
	if len(c.linearizable(ops)) == 0 {
		t.Error("expected operations to NOT be linearizable")
	}

	// read misses a previous write is not linearizable
	// +--w1--+ +--w2--+ +--r1--+
	ops = []*operation{
		&operation{n1, nil, 0, 5},
		&operation{n2, nil, 6, 10},
		&operation{nil, n1, 11, 15},
	}
	if len(c.linearizable(ops)) == 0 {
		t.Error("expected operations to NOT be linearizable")
	}

	// cross reads is not linearizable
	// +--w1--+  +--r1--+
	// +--w2--+  +--r2--+
	ops = []*operation{
		&operation{n1, nil, 0, 5},
		&operation{n2, nil, 0, 5},
		&operation{nil, n1, 6, 10},
		&operation{nil, n2, 6, 10},
	}
	if len(c.linearizable(ops)) == 0 {
		t.Error("expected operations to NOT be linearizable")
	}

	// two amonaly reads
	// +--w1--+ +--w2--+ +--r1--+
	//                     +--r1--+
	ops = []*operation{
		&operation{n1, nil, 0, 5},
		&operation{n2, nil, 6, 10},
		&operation{nil, n1, 11, 15},
		&operation{nil, n1, 12, 16},
	}
	n := len(c.linearizable(ops))
	if n != 2 {
		t.Errorf("expected two amonaly operations, detected %d", n)
	}

	// test link between two writes
	// +--w1--+ +--r1--+ +--r1--+
	//          +--w2--+

	ops = []*operation{
		&operation{n1, nil, 0, 5},
		&operation{nil, n1, 6, 10},
		&operation{n2, nil, 7, 10},
		&operation{nil, n1, 11, 15},
	}
	if len(c.linearizable(ops)) == 0 {
		t.Errorf("expected violation")
	}
}

func TestNonUniqueValue(t *testing.T) {
	c := newChecker()

	// cross read same value should be linearizable
	// +--w1--+  +--r1--+
	// +--w1--+  +--r1--+
	n1 := make([]byte, 4)
	binary.LittleEndian.PutUint32(n1, 1)
	ops := []*operation{
		&operation{n1, nil, 0, 5},
		&operation{n1, nil, 0, 5},
		&operation{nil, n1, 6, 10},
		&operation{nil, n1, 6, 10},
	}

	n := len(c.linearizable(ops))
	if n != 0 {
		t.Errorf("expected no violation, detected %d", n)
	}
}
