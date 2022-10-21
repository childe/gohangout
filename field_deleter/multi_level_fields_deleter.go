package field_deleter

type MultiLevelFieldDeleter struct {
	preFields []string
	lastField string
}

func NewMultiLevelFieldDeleter(fields []string) *MultiLevelFieldDeleter {
	fieldsLength := len(fields)
	preFields := make([]string, fieldsLength-1)
	for i := range preFields {
		preFields[i] = fields[i]
	}

	return &MultiLevelFieldDeleter{
		preFields: preFields,
		lastField: fields[fieldsLength-1],
	}
}

func (d *MultiLevelFieldDeleter) Delete(event map[string]interface{}) {
	current := event
	for _, field := range d.preFields {
		if v, ok := current[field]; ok {
			if current, ok = v.(map[string]interface{}); !ok {
				return
			}
		} else {
			return
		}
	}
	delete(current, d.lastField)
}
