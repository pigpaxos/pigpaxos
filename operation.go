package paxi

import "fmt"

type operation struct {
	input  Value
	output Value
	// timestamps
	start int64
	end   int64
}

func (a operation) happenBefore(b operation) bool {
	return a.end < b.start
}

func (a operation) concurrent(b operation) bool {
	return !a.happenBefore(b) && !b.happenBefore(a)
}

func (a operation) equal(b operation) bool {
	return a.input.Equals(b.input) && a.output.Equals(b.output) && a.start == b.start && a.end == b.end
}

func (a operation) String() string {
	return fmt.Sprintf("{input=%x, output=%x, start=%d, end=%d}", a.input, a.output, a.start, a.end)
}

// sort operations by invocation time
type byTime []*operation

func (a byTime) Len() int           { return len(a) }
func (a byTime) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a byTime) Less(i, j int) bool { return a[i].start < a[j].start }
