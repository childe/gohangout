package value_render

import "github.com/oliveagle/jsonpath"

type JsonpathRender struct {
	Pat *jsonpath.Compiled
}

func (r *JsonpathRender) Render(event map[string]interface{}) (exist bool, value interface{}) {
	if value, err := r.Pat.Lookup(event); err == nil {
		return true, value
	}
	return false, nil
}
