package value_render

import "github.com/oliveagle/jsonpath"

type JsonpathRender struct {
	Pat *jsonpath.Compiled
}

func (r *JsonpathRender) Render(event map[string]interface{}) interface{} {
	if value, ok := r.Pat.Lookup(event); ok == nil {
		return value
	}
	return nil
}
