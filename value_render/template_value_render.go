package value_render

import (
	"text/template"

	"github.com/golang/glog"
)

type TemplateValueRender struct {
	tmpl *template.Template
}

func NewTemplateValueRender(t string) *TemplateValueRender {
	tmpl, err := template.New(t).Parse(t)
	if err != nil {
		glog.Fatalf("could not parse template %s:%s", t, err)
	}
	return &TemplateValueRender{
		tmpl: tmpl,
	}
}

func (r *TemplateValueRender) Render(event map[string]interface{}) interface{} {
	return nil
}
