package gmvc

import (
    "html/template"
    "bytes"
)

type html struct {
    dir   string
    ext   string
    layout string
    funcs template.FuncMap
}

func newHtml() *html {
    h := &html{dir: "template/", ext: "html", layout: "layout/main"}
    h.funcs = template.FuncMap{
        "html":    h.html,
    }
    return h
}

func (h *html) html(str string) template.HTML {
    return template.HTML(str)
}

func (h *html) templateFile(name string) string {
    return h.dir + name + "." + h.ext
}

/**
data
 */
func (h *html) render(data interface{}, tpls ...string) []byte {
    var name, layout string
    l := len(tpls)

    if l == 0 {
        panic("gmvc: missing template name")
    } else if l == 1 {
        name = tpls[0]
        layout = h.layout
    } else {
        name = tpls[0]
        layout = tpls[1]
    }

    tpl := template.Must(template.ParseFiles(h.templateFile(layout), h.templateFile(name))).Funcs(h.funcs)
    buffer := &bytes.Buffer{}
    tpl.Execute(buffer, data)
    return buffer.Bytes()
}

