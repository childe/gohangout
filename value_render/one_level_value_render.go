package value_render

type OneLevelValueRender struct {
	field string
}

func NewOneLevelValueRender(template string) *OneLevelValueRender {
	return &OneLevelValueRender{
		field: template,
	}
}

func (vr *OneLevelValueRender) Render(event map[string]interface{}) interface{} {
	if value, ok := event[vr.field]; ok {
		return value
	}
	return nil
}
