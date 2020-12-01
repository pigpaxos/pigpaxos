package retro_log

import (
	"github.com/pigpaxos/pigpaxos/log"
	"sync"
)

type TimerRqlSet struct {
	name string
	expireTime int64

	bucketTs []int64
	buckets  [][]string

	sync.RWMutex
}

func NewTimerRqlSet(name string, expireTime int64) * TimerRqlSet {
	return &TimerRqlSet{name:name, expireTime:expireTime, buckets: make([][]string, 0), bucketTs: make([]int64, 0)}
}

func (t *TimerRqlSet) Add(val string, ts int64) {
	t.Lock()
	defer t.Unlock()

	if len(t.bucketTs) == 0 {
		t.bucketTs = append(t.bucketTs, ts)
		t.buckets = append(t.buckets, make([]string, 0))
	} else {
		bts := t.bucketTs[len(t.bucketTs)-1]
		if bts != ts {
			t.bucketTs = append(t.bucketTs, ts)
			t.buckets = append(t.buckets, make([]string, 0))
		}
	}
	log.Debugf("Adding Items to bucket %d", ts)
	t.buckets[len(t.buckets) - 1] = append(t.buckets[len(t.buckets) - 1], val)

}

func (t *TimerRqlSet) ExpireItems(ts int64) *RqlSet {
	t.Lock()
	defer t.Unlock()

	removeSet := NewRqlSet(t.name, false)

	lastExpiredIndex := -1
	for index, bts := range t.bucketTs {
		if bts + t.expireTime <= ts {
			log.Debugf("Expiring Items in bucket %d at time %d", bts, ts)
			removeSet.addAllVal(t.buckets[index])
			lastExpiredIndex = index
		}
	}

	if lastExpiredIndex >= 0 {
		t.bucketTs = append(t.bucketTs[:0], t.bucketTs[lastExpiredIndex + 1:]...)
		t.buckets = append(t.buckets[:0], t.buckets[lastExpiredIndex + 1:]...)
	}

	return removeSet
}

func (t *TimerRqlSet) Snapshot() *RqlSet{
	appendSet := NewRqlSet(t.name, true)
	for _, bucket := range t.buckets {
		appendSet.addAllVal(bucket)
	}
	return appendSet
}