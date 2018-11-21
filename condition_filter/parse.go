package condition_filter

import (
	"errors"
	"fmt"
	"strings"
)

const (
	OP_NONE = iota
	OP_NOT
	OP_AND
	OP_OR
)

const (
	PARSE_OUTSIDE_CONDITION = iota
	PARSE_STATE_IN_ONE_CONDITION
	PARSE_STATE_IN_STRING
)

func process_node_stack(node_stack []*OPNode, force bool) ([]*OPNode, error) {
	length := len(node_stack)
	if length == 0 {
		return node_stack, nil
	}

	last_node := node_stack[length-1]
	if length == 1 {
		if last_node.left == nil && last_node.condition == nil {
			if last_node.op == OP_AND || last_node.op == OP_OR {
				return node_stack, errors.New("it is illegal that first element is && or ||")
			}
		}
		return node_stack, nil
	}

	last_2_node := node_stack[length-2]
	if last_node.op == OP_NONE || last_node.left != nil { // this is not pure oper node

		switch last_2_node.op {
		case OP_NOT:
			last_2_node.left = last_node
			node_stack = node_stack[:length-1]
			return process_node_stack(node_stack, force)
		case OP_AND:
			last_2_node.left = node_stack[length-3]
			last_2_node.right = last_node
			node_stack = append(node_stack[:length-3], last_2_node)
			return process_node_stack(node_stack, force)
		case OP_NONE:
			return node_stack, errors.New("2 successive condition (no && || between them) is illegal")
		case OP_OR:
			if !force {
				return node_stack, nil
			}
			last_2_node.left = node_stack[length-3]
			last_2_node.right = last_node
			node_stack = append(node_stack[:length-3], last_2_node)
			return process_node_stack(node_stack, force)
		}
	}

	// last node is pure oper node
	switch last_node.op {
	case OP_NOT:
		return node_stack, nil
	case OP_AND:
		return node_stack, nil
	case OP_OR:
		new_stack, err := process_node_stack(node_stack[:length-1], true)
		if err != nil {
			return node_stack, nil
		}
		return append(new_stack, last_node), nil
	}

	return node_stack, nil

	//  ========= NO check ==========

	/*** check ***

	if last_node.left != nil {
		if last_2_node.left != nil {
			return errors.New("2 successive condition is illegal")
		}
		if last_2_node.op == OP_NOT {
		}
	}

	switch last_node.op {
	case OP_NONE:
		if last_2_node.op == OP_NONE {
			return errors.New("2 successive condition is illegal")
		}
		if last_2_node.op == OP_NOT {
			last_2_node.left = last_node
			node_stack = node_stack[:length-1]
			return process_node_stack(node_stack)
		}

		if last_2_node.op == OP_AND {
			if length < 3 {
				return errors.New("should has one condition before &&")
			}
			last_2_node.left = node_stack[lenght-2]
			last_2_node.right = last_node
			node_stack = node_stack[:length-2]
			return process_node_stack(node_stack)
		}

		if last_2_node.op == OP_OR {
			if force == false {
				return nil
			}
			if length < 3 {
				return errors.New("should has one condition before ||")
			}
			last_2_node.left = node_stack[lenght-2]
			last_2_node.right = last_node
			node_stack = node_stack[:length-2]
			return process_node_stack(node_stack, false)
		}
	case OP_NOT:
		if last_2_node.op == OP_NONE {
			return errors.New("it is illegal that `!` is after direct condition")
		}
	case OP_AND:
		if last_2_node.op == OP_AND || last_2_node.op == OP_OR {
			return errors.New("2 successive && or || is illegal")
		}
	case OP_OR:
		if last_2_node.op == OP_AND || last_2_node.op == OP_OR {
			return errors.New("2 successive && or || is illegal")
		}
		new_stack := process_node_stack(node_stack[:length-1], true)
	}

	return nil
	*/
}

func check_node_stack(node_stack []*OPNode) (*OPNode, error) {
	return node_stack[0], nil
}

func parseBoolTree(c string) (*OPNode, error) {
	c = strings.Trim(c, " ")
	if c == "" {
		return nil, nil
	}

	if c[0] == '|' || c[0] == '&' {
		return nil, errors.New("condition could not starts with & or |")
	}

	var (
		i                   int = 0
		length              int = len(c)
		state               int = PARSE_OUTSIDE_CONDITION
		bracke_stack        int
		op                  int       = OP_NONE
		condition_start_pos int       = 0
		node_stack          []*OPNode = make([]*OPNode, 0)
		err                 error
		//last_element_type   int = OP_NONE
	)

	for i < length {
		if state == PARSE_OUTSIDE_CONDITION {
			for ; i < length; i++ {
				if c[i] == ' ' {
					continue
				} else if c[i] == '!' {
					op = OP_NOT
					n := &OPNode{
						op:        op,
						left:      nil,
						right:     nil,
						condition: nil,
					}
					node_stack = append(node_stack, n)
					i++
					condition_start_pos = i
					break
				} else if c[i] == '|' {
					if c[i+1] != '|' {
						return nil, fmt.Errorf("column %d illegal |", i)
					}
					op = OP_OR
					n := &OPNode{
						op:        op,
						left:      nil,
						right:     nil,
						condition: nil,
					}
					node_stack = append(node_stack, n)
					i += 2
					condition_start_pos = i
					break
				} else if c[i] == '&' {
					if c[i+1] != '&' {
						return nil, fmt.Errorf("column %d illegal &", i)
					}
					op = OP_AND
					n := &OPNode{
						op:        op,
						left:      nil,
						right:     nil,
						condition: nil,
					}
					node_stack = append(node_stack, n)
					i += 2
					condition_start_pos = i
					break
				} else {
					state = PARSE_STATE_IN_ONE_CONDITION
					condition_start_pos = i
					break
				}
			}
		} else if state == PARSE_STATE_IN_STRING {
			for ; i < length; i++ {
				if c[i] == '"' {
					state = PARSE_STATE_IN_ONE_CONDITION
					i++
					break
				}
			}
		} else if state == PARSE_STATE_IN_ONE_CONDITION {
			for ; i < length; i++ {
				if c[i] == '"' {
					state = PARSE_STATE_IN_STRING
					i++
					break
				}
				if c[i] == '(' {
					bracke_stack++
				}
				if c[i] == ')' {
					bracke_stack--
					if bracke_stack == 0 {
						n := &OPNode{
							op:        OP_NONE,
							left:      nil,
							right:     nil,
							condition: NewCondition(c[condition_start_pos : i+1]),
						}
						node_stack = append(node_stack, n)
						state = PARSE_OUTSIDE_CONDITION
						i++
						break
					}
				}
			}
		}

		if node_stack, err = process_node_stack(node_stack, false); err != nil {
			return nil, err
		}

	}

	if len(node_stack) == 0 {
		return nil, errors.New("unclosed condition")
	}

	if state != PARSE_OUTSIDE_CONDITION {
		return nil, errors.New("unclosed condition")
	}

	if node_stack, err = process_node_stack(node_stack, true); err != nil {
		return nil, err
	} else {
		return check_node_stack(node_stack)
	}
}
