package main

type InputBox struct {
	input   Input
	filters []Filter
	outputs []Output
	config  map[interface{}]interface{}
}

func (box *InputBox) prepare(event map[string]interface{}) map[string]interface{} {
	return event
}

func (box *InputBox) beat() {
	//box.input.init(box.config)

	for {
		event := box.input.readOneEvent()
		if typeValue, ok := box.config["type"]; ok {
			event["type"] = typeValue
		}
		if box.filters != nil {
			for _, filterPlugin := range box.filters {
				event = filterPlugin.process(event)
			}
		}
		for _, outputPlugin := range box.outputs {
			outputPlugin.emit(event)
		}
	}
}
