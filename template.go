package gmvc

import (
    "html/template"
    "bytes"
)

var (
    TemplateExt = ".html"
)

type Template struct {
    dir   string
    funcs template.FuncMap
}

func NewTemplate() *Template {
    t := &Template{dir: "templates/"}
    t.funcs = template.FuncMap{
        "html":    Html,
    }
    return t
}

func (t *Template) AddFuncs(funcs map[string]interface{}) {
    for k, f := range funcs {
        t.funcs[k] = f
    }
}

func (t *Template) SetRootDir(dir string) {
    t.dir = dir
}

func (t *Template) templateFile(name string) string {
    return t.dir + "/" + name + TemplateExt
}

func (t *Template) render(layout, name string, data interface{}) []byte {
    tpl := template.New("/")
    template.Must(tpl.Funcs(t.funcs).ParseFiles(t.templateFile(layout), t.templateFile(name)))
    buffer := &bytes.Buffer{}
    tpl.Execute(buffer, data)
    return buffer.Bytes()
}

func Html(str string) template.HTML {
    return template.HTML(str)
}

