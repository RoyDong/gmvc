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
    h := &html{dir: "templates/", ext: "html", layout: "layouts/main"}
    h.funcs = template.FuncMap{
        "html":    Html,
    }
    return h
}

func (h *html) templateFile(name string) string {
    return h.dir + "/" + name + "." + h.ext
}

/**
data
 */
func (h *html) render(data interface{}, tpls ...string) []byte {
    var name, layout string
    l := len(tpl)

    if l == 0 {
        panic("gmvc: missing template name")
    } else if l == 1 {
        name = tpl[0]
        layout = h.layout
    } else {
        name = tpl[0]
        layout = tpl[1]
    }

    tpl := template.New("/").Funcs(h.funcs)
    template.Must(tpl.Funcs(h.funcs).ParseFiles(h.templateFile(layout), h.templateFile(name)))
    buffer := &bytes.Buffer{}
    tpl.Execute(buffer, data)
    return buffer.Bytes()
}

