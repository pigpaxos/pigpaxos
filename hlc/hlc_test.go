package hlc

import (
	"testing"
	"github.com/stretchr/testify/assert"
	"time"
	"github.com/acharapko/fleetdb/utils"
)

func TestNewHLC(t *testing.T) {
	pt := time.Now().Unix()
	hlc := NewHLC(pt)
	assert.Equal(t, pt, hlc.currentHLC.PhysicalTime, "expected %d physical time, got %d", pt, hlc.currentHLC.PhysicalTime)
	assert.Equal(t, int16(0), hlc.currentHLC.LogicalTime, "expected 0 logical time, got %d", hlc.currentHLC.LogicalTime)
}

func TestEmptyHLC(t *testing.T) {
	hlc := NewHLC(0)
	assert.Equal(t, int64(0), hlc.currentHLC.PhysicalTime, "expected %d physical time, got %d", 0, hlc.currentHLC.PhysicalTime)
	assert.Equal(t, int16(0), hlc.currentHLC.LogicalTime, "expected 0 logical time, got %d", hlc.currentHLC.LogicalTime)
}


func TestHLCNow(t *testing.T) {
	pt := utils.CurrentTimeInMS()
	hlc := NewHLC(pt)
	assert.Equal(t, pt, hlc.currentHLC.PhysicalTime, "expected %d physical time, got %d", pt, hlc.currentHLC.PhysicalTime)
	assert.Equal(t, int16(0), hlc.currentHLC.LogicalTime, "expected 0 logical time, got %d", hlc.currentHLC.LogicalTime)
	time.Sleep(10 * time.Millisecond)
	hlc.Now()
	assert.Less(t, pt, hlc.currentHLC.PhysicalTime, "expected pt increment")
	assert.Equal(t, int16(0), hlc.currentHLC.LogicalTime, "expected 0 logical time, got %d", hlc.currentHLC.LogicalTime)

	// HLC with higher PT
	pt = utils.CurrentTimeInMS() + 100
	hlc = NewHLC(pt)
	assert.Equal(t, pt, hlc.currentHLC.PhysicalTime, "expected %d physical time, got %d", pt, hlc.currentHLC.PhysicalTime)
	assert.Equal(t, int16(0), hlc.currentHLC.LogicalTime, "expected 0 logical time, got %d", hlc.currentHLC.LogicalTime)
	time.Sleep(10 * time.Millisecond)
	hlc.Now()
	assert.Equal(t, pt, hlc.currentHLC.PhysicalTime, "expected %d physical time, got %d", pt, hlc.currentHLC.PhysicalTime)
	assert.Equal(t, int16(1), hlc.currentHLC.LogicalTime, "expected 1 logical time, got %d", hlc.currentHLC.LogicalTime)
}

func TestHLCUpdate(t *testing.T) {
	pt := utils.CurrentTimeInMS()
	hlc1 := NewHLC(pt)
	assert.Equal(t, pt, hlc1.currentHLC.PhysicalTime, "expected %d physical time, got %d", pt, hlc1.currentHLC.PhysicalTime)
	assert.Equal(t, int16(0), hlc1.currentHLC.LogicalTime, "expected 0 logical time, got %d", hlc1.currentHLC.LogicalTime)

	pt = utils.CurrentTimeInMS() + 100
	hlc2 := NewHLC(pt)
	assert.Equal(t, pt, hlc2.currentHLC.PhysicalTime, "expected %d physical time, got %d", pt, hlc2.currentHLC.PhysicalTime)
	assert.Equal(t, int16(0), hlc2.currentHLC.LogicalTime, "expected 0 logical time, got %d", hlc2.currentHLC.LogicalTime)

	hlc1.Update(*hlc2.currentHLC)

	// now we expect hlc1.pt time to match hlc2.pt, and hlc2.lc to increment
	assert.Equal(t, hlc2.currentHLC.PhysicalTime, hlc1.currentHLC.PhysicalTime, "expected %d physical time, got %d", hlc2.currentHLC.PhysicalTime, hlc1.currentHLC.PhysicalTime)
	assert.Equal(t, int16(1), hlc1.currentHLC.LogicalTime, "expected %d logical time, got %d", hlc2.currentHLC.LogicalTime, hlc1.currentHLC.LogicalTime)

	pt = hlc2.currentHLC.PhysicalTime - 100
	hlc3 := NewHLC(pt)
	assert.Equal(t, pt, hlc3.currentHLC.PhysicalTime, "expected %d physical time, got %d", pt, hlc3.currentHLC.PhysicalTime)
	assert.Equal(t, int16(0), hlc3.currentHLC.LogicalTime, "expected 0 logical time, got %d", hlc3.currentHLC.LogicalTime)

	ptCheck := hlc1.currentHLC.PhysicalTime
	hlc1.Update(*hlc3.currentHLC)
	// now we expect HLC1 time to tick logical
	assert.Less(t, hlc3.currentHLC.PhysicalTime, hlc1.currentHLC.PhysicalTime)
	assert.Equal(t, ptCheck, hlc1.currentHLC.PhysicalTime)
	assert.Equal(t, int16(2), hlc1.currentHLC.LogicalTime, "expected 1 logical time, got %d", hlc1.currentHLC.LogicalTime)

	// same HLC tme
	pt = hlc1.currentHLC.PhysicalTime
	hlc4 := NewHLC(pt)
	assert.Equal(t, pt, hlc4.currentHLC.PhysicalTime, "expected %d physical time, got %d", pt, hlc4.currentHLC.PhysicalTime)
	assert.Equal(t, int16(0), hlc4.currentHLC.LogicalTime, "expected 0 logical time, got %d", hlc4.currentHLC.LogicalTime)

	hlc1.Update(*hlc4.currentHLC)
	// now we expect HLC1 time to tick logical
	assert.Equal(t, hlc4.currentHLC.PhysicalTime, hlc1.currentHLC.PhysicalTime)
	assert.Equal(t, int16(3), hlc1.currentHLC.LogicalTime, "expected 1 logical time, got %d", hlc1.currentHLC.LogicalTime)
}