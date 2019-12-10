package topology

type ProcessorInLink interface {
	// will process event in filter/output , and the call nexter.process, until nexter is nil
	Process(map[string]interface{}) map[string]interface{}
}

type NilProcessorInLink struct{}

func (n *NilProcessorInLink) Process(event map[string]interface{}) map[string]interface{} {
	return event
}
