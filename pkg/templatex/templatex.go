package templatex

import (
	"github.com/valyala/fasttemplate"
)

// Template 封装了fasttemplate.Template
type Template struct {
	origin   string
	template *fasttemplate.Template
}

func New(s string) *Template {
	return &Template{template: fasttemplate.New(s, "{", "}"), origin: s}
}

// Execute 创建一个Render并返回
func (t *Template) Execute(vals map[string]any) string {
	return t.template.ExecuteString(vals)
}
