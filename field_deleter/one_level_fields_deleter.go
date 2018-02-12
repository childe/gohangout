package field_deleter

type OneLevelFieldDeleter struct {
	field string
}

func NewOneLevelFieldDeleter(field string) *OneLevelFieldDeleter {
	return &OneLevelFieldDeleter{
		field: field,
	}
}

func (d *OneLevelFieldDeleter) Delete(event map[string]interface{}) {
	delete(event, d.field)
}
