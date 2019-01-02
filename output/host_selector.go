package output

import (
	"math/rand"
	"time"
)

type HostSelector interface {
	Next() interface{}
	ReduceWeight()
	AddWeight()
	Size() int
}

type RRHostSelector struct {
	hosts      []interface{}
	initWeight int
	weight     []int
	index      int
	hostsCount int
}

func NewRRHostSelector(hosts []interface{}, weight int) *RRHostSelector {
	rand.Seed(time.Now().UnixNano())
	hostsCount := len(hosts)
	rst := &RRHostSelector{
		hosts:      hosts,
		index:      int(rand.Int31n(int32(hostsCount))),
		hostsCount: hostsCount,
		initWeight: weight,
	}
	rst.weight = make([]int, hostsCount)
	for i := 0; i < hostsCount; i++ {
		rst.weight[i] = weight
	}

	return rst
}

func (s *RRHostSelector) Next() interface{} {
	// reset weight and return "" if all hosts are down
	var hasAtLeastOneUp bool = false
	for i := 0; i < s.hostsCount; i++ {
		if s.weight[i] > 0 {
			hasAtLeastOneUp = true
		}
	}
	if !hasAtLeastOneUp {
		s.resetWeight(s.initWeight)
		return nil
	}

	s.index = (s.index + 1) % s.hostsCount
	return s.hosts[s.index]
}

func (s *RRHostSelector) resetWeight(weight int) {
	for i := range s.weight {
		s.weight[i] = weight
	}
}

func (s *RRHostSelector) ReduceWeight() {
	if s.weight[s.index] > 0 {
		s.weight[s.index]--
	}
}

func (s *RRHostSelector) AddWeight() {
	s.weight[s.index] = s.weight[s.index] + 1
	if s.weight[s.index] > s.initWeight {
		s.weight[s.index] = s.initWeight
	}
}

func (s *RRHostSelector) Size() int {
	return len(s.hosts)
}
