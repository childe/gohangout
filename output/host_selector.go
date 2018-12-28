package output

import (
	"math/rand"
	"time"
)

type HostSelector interface {
	selectOneHost() string
	reduceWeight(string)
	addWeight(string)
}

type RRHostSelector struct {
	hosts      []string
	initWeight int
	weight     []int
	index      int
	hostsCount int
}

func NewRRHostSelector(hosts []string, weight int) *RRHostSelector {
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

func (s *RRHostSelector) selectOneHost() string {
	// reset weight and return "" if all hosts are down
	var hasAtLeastOneUp bool = false
	for i := 0; i < s.hostsCount; i++ {
		if s.weight[i] > 0 {
			hasAtLeastOneUp = true
		}
	}
	if !hasAtLeastOneUp {
		s.resetWeight(s.initWeight)
		return ""
	}

	s.index = (s.index + 1) % s.hostsCount
	return s.hosts[s.index]
}

func (s *RRHostSelector) resetWeight(weight int) {
	for i := range s.weight {
		s.weight[i] = weight
	}
}

func (s *RRHostSelector) reduceWeight(host string) {
	for i, h := range s.hosts {
		if host == h {
			s.weight[i] = s.weight[i] - 1
			if s.weight[i] < 0 {
				s.weight[i] = 0
			}
			return
		}
	}
}

func (s *RRHostSelector) addWeight(host string) {
	for i, h := range s.hosts {
		if host == h {
			s.weight[i] = s.weight[i] + 1
			if s.weight[i] > s.initWeight {
				s.weight[i] = s.initWeight
			}
			return
		}
	}
}
