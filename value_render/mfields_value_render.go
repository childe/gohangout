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
func (vr *MultiLevelValueRender) Render(event map[string]interface{}) (exist bool, value interface{}) {
	var current map[string]interface{} = event
	for _, field := range vr.preFields {
		value, exist = current[field]
		if !exist {
			return false, nil
		}
		if value == nil {
			return true, nil
		}
		if current, exist = value.(map[string]interface{}); !exist {
			return false, nil
		}
	}

	if value, ok := current[vr.lastField]; ok {
		return true, value
	} else {
		return false, nil
	}
}
