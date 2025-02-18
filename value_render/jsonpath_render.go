package value_render

import "github.com/oliveagle/jsonpath"

type JsonpathRender struct {
	Pat *jsonpath.Compiled
}

func (r *JsonpathRender) Render(event map[string]interface{}) (value interface{}, err error) {
	return r.Pat.Lookup(event)
}
