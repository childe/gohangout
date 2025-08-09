package field_setter

type OneLevelFieldSetter struct {
	field string
}

func NewOneLevelFieldSetter(field string) *OneLevelFieldSetter {
	r := &OneLevelFieldSetter{
		field: field,
	}
	return r
}

func (fs *OneLevelFieldSetter) SetField(event map[string]any, value any, field string, overwrite bool) map[string]any {
	if _, ok := event[fs.field]; !ok || overwrite {
		event[fs.field] = value
	}
	return event
}
