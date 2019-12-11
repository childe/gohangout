package topology

// FilterBox and OutputBox is Processor
type Processor interface {
	Process(map[string]interface{}) map[string]interface{}
}

type NilProcessorInLink struct{}

func (n *NilProcessorInLink) Process(event map[string]interface{}) map[string]interface{} {
	return event
}

// ProcessorNode is a node in the filter/output link
type ProcessorNode struct {
	Processor Processor
	Next      *ProcessorNode
}

// Processor will process event , and pass it to next, and then next , until last one(generally output)
func (node *ProcessorNode) Process(event map[string]interface{}) map[string]interface{} {
	event = node.Processor.Process(event)
	if event == nil || node.Next == nil {
		return event
	}

	return node.Next.Process(event)
}

// AppendProcessorsToLink add new processors to tail, return head node
func AppendProcessorsToLink(head *ProcessorNode, processors ...Processor) *ProcessorNode {
	preHead := &ProcessorNode{nil, head}
	n := preHead

	// look for tail
	for n.Next != nil {
		n = n.Next
	}

	for _, processor := range processors {
		n.Next = &ProcessorNode{processor, nil}
		n = n.Next
	}

	return preHead.Next
}
