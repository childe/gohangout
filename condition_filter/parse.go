package condition_filter

import (
	"errors"
	"strings"

	"github.com/golang/glog"
)

const (
	_op_sharp = iota
	_op_left
	_op_right
	_op_or
	_op_and
	_op_not
)

const (
	_OUTSIDES_CONDITION = iota
	_IN_CONDITION
	_IN_STRING
)

var errorParse = errors.New("parse condition error")

func parseBoolTree(c string) (node *OPNode, err error) {
	defer func() {
		if r := recover(); r != nil {
			glog.Errorf("parse `%s` error at `%s`", c, r)
			node = nil
			err = errorParse
		}
	}()

	//glog.Info(c)
	c = strings.Trim(c, " ")
	if c == "" {
		return nil, nil
	}

	s2, err := buildRPNStack(c)
	if err != nil {
		return nil, err
	}
	//glog.Info(s2)
	s := make([]interface{}, 0)

	for _, e := range s2 {
		if c, ok := e.(Condition); ok {
			s = append(s, c)
		} else {
			sLen := len(s)
			op := e.(int)
			if op == _op_not {
				right := s[sLen-1].(*OPNode)
				s = s[:sLen-1]
				node := &OPNode{
					op:    op,
					right: right,
				}
				s = append(s, node)
			} else {
				right := s[sLen-1].(*OPNode)
				left := s[sLen-2].(*OPNode)
				s = s[:sLen-2]
				node := &OPNode{
					op:    op,
					left:  left,
					right: right,
				}
				s = append(s, node)
			}
		}
	}

	//glog.Info(s)
	if len(s) != 1 {
		return nil, errorParse
	}
	return s[0].(*OPNode), nil
}

func buildRPNStack(c string) ([]interface{}, error) {
	var (
		state               = _OUTSIDES_CONDITION
		i                   int
		length              = len(c)
		parenthesis         = 0
		condition_start_pos int

		s1 = []int{_op_sharp}
		s2 = make([]interface{}, 0)
	)

	// 哪些导致状态变化??

	for i < length {
		switch c[i] {
		case '(':
			switch state {
			case _OUTSIDES_CONDITION: // push s1
				s1 = append(s1, _op_left)
			case _IN_CONDITION:
				parenthesis++
			}
		case ')':
			switch state {
			case _OUTSIDES_CONDITION:
				if !pushOp(_op_right, &s1, &s2) {
					panic(c[:i+1])
				}

			case _IN_CONDITION:
				parenthesis--
				if parenthesis == 0 {
					condition, err := NewSingleCondition(c[condition_start_pos : i+1])
					if err != nil {
						glog.Error(err)
						panic(c[:i+1])
					}
					n := &OPNode{
						condition: condition,
					}
					s2 = append(s2, n)
					state = _OUTSIDES_CONDITION
				}
			}
		case '&':
			switch state {
			case _OUTSIDES_CONDITION: // push s1
				if c[i+1] != '&' {
					panic(c[:i+1])
				} else {
					if !pushOp(_op_and, &s1, &s2) {
						panic(c[:i+1])
					}
					i++
				}
			}
		case '|':
			switch state {
			case _OUTSIDES_CONDITION: // push s1
				if c[i+1] != '|' {
					panic(c[:i+1])
				} else {
					if !pushOp(_op_or, &s1, &s2) {
						panic(c[:i+1])
					}
					i++
				}
			}
		case '!':
			switch state {
			case _OUTSIDES_CONDITION: // push s1
				if n := c[i+1]; n == '|' || n == '&' || n == ' ' {
					panic(c[:i+1])
				}
				if !pushOp(_op_not, &s1, &s2) {
					panic(c[:i+1])
				}
			}
		case '"':
			switch state {
			case _OUTSIDES_CONDITION: // push s1
				panic(c[:i+1])
			case _IN_STRING:
				state = _IN_CONDITION
			}
		case ' ':
		default:
			if state == _OUTSIDES_CONDITION {
				state = _IN_CONDITION
				condition_start_pos = i
			}

		}
		i++
	}

	if state != _OUTSIDES_CONDITION {
		return nil, errorParse
	}

	for j := len(s1) - 1; j > 0; j-- {
		s2 = append(s2, s1[j])
	}

	return s2, nil
}

func pushOp(op int, s1 *[]int, s2 *[]interface{}) bool {
	if op == _op_right {
		return findLeftInS1(s1, s2)
	}
	return compareOpWithS1(op, s1, s2)
}

// find ( in s1
func findLeftInS1(s1 *[]int, s2 *[]interface{}) bool {
	var j int
	for j = len(*s1) - 1; j > 0 && (*s1)[j] != _op_left; j-- {
		*s2 = append(*s2, (*s1)[j])
	}

	if j == 0 {
		return false
	}

	*s1 = (*s1)[:j]
	return true
}

// compare op with ops in s1, and put them to s2
func compareOpWithS1(op int, s1 *[]int, s2 *[]interface{}) bool {
	var j int
	for j = len(*s1) - 1; j > 0; j-- {
		//if (*s1)[j] == _op_left || op > (*s1)[j] {
		n1 := (*s1)[j]
		b := true
		switch {
		case n1 == _op_left:
			break
		case op > n1:
			break
		case op == _op_not && n1 == _op_not:
			break
		default:
			b = false
		}
		if b {
			break
		}
		*s2 = append(*s2, n1)
	}

	*s1 = (*s1)[:j+1]
	*s1 = append(*s1, op)
	return true
}
