package main

type LiteralValueRender struct {
	value string
}

func (vr *LiteralValueRender) render(event map[string]interface{}) interface{} {
	return vr.value
}

func NewLiteralValueRender(template string) *LiteralValueRender {
	return &LiteralValueRender{template}
}
