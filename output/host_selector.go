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
	for i := 1; i <= s.hostsCount; i++ {
		idx := (s.index + i) % s.hostsCount
		if s.weight[idx] > 0 {
			s.index = idx
			return s.hosts[idx]
		}
	}

	s.resetWeight(s.initWeight)
	// allow client wait for some time and then get Next
	return nil
}

func (s *RRHostSelector) resetWeight(weight int) {
	for i := range s.weight {
		s.weight[i] = weight
	}
}

func (s *RRHostSelector) ReduceWeight() {
	s.weight[s.index]--
	if s.weight[s.index] <= 0 {
		i := s.index
		time.AfterFunc(time.Minute*30, func() {
			s.weight[i] = 1
		})
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
