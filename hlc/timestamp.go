package hlc

import (
	"math"
	"time"
	"encoding/binary"
)

// Timestamp constant values.
var (
	// MaxTimestamp is the max value allowed for Timestamp.
	MaxTimestamp = Timestamp{PhysicalTime: math.MaxInt64, LogicalTime: math.MaxInt16}
	// MinTimestamp is the min value allowed for Timestamp.
	MinTimestamp = Timestamp{PhysicalTime: 0, LogicalTime: 0}
)

type Timestamp struct {
	PhysicalTime int64
	LogicalTime  int16
}

/* Converts timestamp into Int64 representation of HLC */
func (t Timestamp) ToInt64() int64 {
	i64hlc := t.PhysicalTime
	i64hlc = i64hlc << 16                  //get some room for lc
	i64hlc = i64hlc | int64(t.LogicalTime) //slap lc to last 16 bits which we cleared
	return i64hlc
}

func (t Timestamp) ToBytes() []byte {
	tint := t.ToInt64()

	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, uint64(tint))

	return b
}

/* Creates HLC from I64 representation of HLC time */
func NewTimestampI64(hlc int64) *Timestamp {
	wallTime := hlc >> 16;
	logical := int16(hlc & 0x00000000000000FFFF);
	t := Timestamp{PhysicalTime:wallTime, LogicalTime:logical}
	return &t
}

func NewTimestampBytes(ts []byte) *Timestamp {
	tint := int64(binary.LittleEndian.Uint64(ts))
	return NewTimestampI64(tint)
}

func NewTimestampPt(pt int64) *Timestamp {
	t := Timestamp{PhysicalTime:pt, LogicalTime:0}
	return &t
}

func NewTimestamp(pt int64, lc int16) *Timestamp {
	t := Timestamp{PhysicalTime:pt, LogicalTime:lc}
	return &t
}

func (t Timestamp) GetPhysicalTime() int64 {
	return t.PhysicalTime
}

func (t Timestamp) GetLogicalTime() int16 {
	return t.LogicalTime
}

func (t *Timestamp) IncrementLogical() {
	t.LogicalTime++
}

func (t *Timestamp) ResetLogical() {
	t.LogicalTime = 0
}

func (t *Timestamp) SetPhysicalTime(pt int64) {
	t.PhysicalTime = pt
}

func (t *Timestamp) SetLogicalTime(lc int16) {
	t.LogicalTime = lc;
}

func (t *Timestamp) Compare(t2 *Timestamp) int {
	if t.PhysicalTime > t2.PhysicalTime {
		return 1
	}
	if t.PhysicalTime < t2.PhysicalTime {
		return -1
	}
	if t.LogicalTime > t2.LogicalTime {
		return 1
	}
	if t.LogicalTime < t2.LogicalTime {
		return -1
	}

	return 0
}

// GoTime converts the timestamp to a time.Time.
func (t Timestamp) GoTime() time.Time {
	return time.Unix(0, t.PhysicalTime)
}