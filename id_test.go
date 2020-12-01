package paxi

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestID(t *testing.T) {
	id := NewID(1, 2)
	assert.Equal(t, 1, id.Zone())
	assert.Equal(t, 2, id.Node())

	id2 := NewID(2, 3)
	assert.Equal(t, 2, id2.Zone())
	assert.Equal(t, 3, id2.Node())
}

func TestIDFromString(t *testing.T) {
	id := NewIDFromString("1.2")
	assert.Equal(t, 1, id.Zone())
	assert.Equal(t, 2, id.Node())

	id2 := NewIDFromString("2.3")
	assert.Equal(t, 2, id2.Zone())
	assert.Equal(t, 3, id2.Node())
}

func BenchmarkNewID(b *testing.B) {
	for i := 0; i < b.N; i++ {
		NewID(1, 2)
	}
}

func BenchmarkGetNode(b *testing.B) {
	id := NewID(1, 2)
	for i := 0; i < b.N; i++ {
		id.Node()
	}
}