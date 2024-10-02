package tracker

import (
	"sync/atomic"
)

type (
	UpstreamRoundRobin struct {
		values []*Upstream
		nitems int
		index  int32
	}
)

func NewUpstreamRoundRobin(values []*Upstream) *UpstreamRoundRobin {
	return &UpstreamRoundRobin{
		values: values,
		nitems: len(values),
	}
}

func (rr *UpstreamRoundRobin) Next() *Upstream {
	n := atomic.AddInt32(&rr.index, 1)
	return rr.values[int(n-1)%rr.nitems]
}

func (rr *UpstreamRoundRobin) Len() int {
	return rr.nitems
}

func (rr *UpstreamRoundRobin) Add(n *Upstream) {
	// not thread safe, lock something before calling this
	rr.values = append(rr.values, n)
	rr.nitems++
}

func (rr *UpstreamRoundRobin) Remove(u *Upstream) {
	// not thread safe, lock something before calling this
	itemPos := -1
	for i := range rr.values {
		if u == rr.values[i] {
			itemPos = i
			break
		}
	}

	if itemPos >= 0 {
		rr.values[itemPos] = rr.values[len(rr.values)-1]
		rr.values = rr.values[:len(rr.values)-1]
		rr.nitems--
	}
}
