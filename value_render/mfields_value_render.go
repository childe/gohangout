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
func (vr *MultiLevelValueRender) Render(event map[string]any) (value any, err error) {
	var current map[string]any = event
	for _, field := range vr.preFields {
		v, ok := current[field]
		if !ok {
			return nil, ErrNotExist
		}
		if current, ok = v.(map[string]any); !ok {
			return nil, ErrInvalidType
		}
	}

	if v, ok := current[vr.lastField]; ok {
		return v, nil
	} else {
		return nil, ErrNotExist
	}
}
