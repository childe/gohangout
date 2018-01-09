package value_render

type LiteralValueRender struct {
	value string
}

func NewLiteralValueRender(template string) *LiteralValueRender {
	return &LiteralValueRender{template}
}

func (r *LiteralValueRender) Render(event map[string]interface{}) interface{} {
	return r.value
}
