package value_render

// MultiLevelValueRender is a ValueRender that can render [xxx][yyy][zzz]
type MultiLevelValueRender struct {
	preFields []string
	lastField string
}

// NewMultiLevelValueRender create a MultiLevelValueRender
func NewMultiLevelValueRender(fields []string) *MultiLevelValueRender {
	fieldsLength := len(fields)
	preFields := make([]string, fieldsLength-1)
	for i := range preFields {
		preFields[i] = fields[i]
	}

	return &MultiLevelValueRender{
		preFields: preFields,
		lastField: fields[fieldsLength-1],
	}
}

// Render implements ValueRender
func (vr *MultiLevelValueRender) Render(event map[string]interface{}) interface{} {
	var current map[string]interface{} = event
	var value interface{}
	var ok bool
	for _, field := range vr.preFields {
		value, ok = current[field]
		if !ok || value == nil {
			return nil
		}
		if current, ok = value.(map[string]interface{}); !ok {
			return nil
		}
	}
	if value, ok := current[vr.lastField]; ok {
		return value
	}
	return nil
}
