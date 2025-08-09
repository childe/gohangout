package value_render

import "github.com/oliveagle/jsonpath"

type JsonpathRender struct {
	Pat *jsonpath.Compiled
}

func (r *JsonpathRender) Render(event map[string]any) (value any, err error) {
	return r.Pat.Lookup(event)
}
