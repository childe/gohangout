package value_render

type OneLevelValueRender struct {
	field string
}

func NewOneLevelValueRender(template string) *OneLevelValueRender {
	return &OneLevelValueRender{
		field: template,
	}
}

func (vr *OneLevelValueRender) Render(event map[string]interface{}) (exist bool, value interface{}) {
	if value, ok := event[vr.field]; ok {
		return true, value
	}
	return false, nil
}
