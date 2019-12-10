package topology

type ProcesserInLink interface {
	// will process event in filter/output , and the call nexter.process, until nexter is nil
	Process(map[string]interface{}) map[string]interface{}
}

type NilProcesserInLink struct{}

func (n *NilProcesserInLink) Process(event map[string]interface{}) map[string]interface{} {
	return event
}
