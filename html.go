package gmvc

import (
    "html/template"
    "bytes"
    "fmt"
    "os"
)

type html struct {
    dir   string
    ext   string
    layout string
    funcs template.FuncMap
}

func newHtml() *html {
    h := &html{layout: fmt.Sprintf("layout%vmain", os.PathSeparator)}

    conf := Store.Tree("config.template")
    var has bool
    h.ext, has = conf.String("ext")
    if !has {
        h.ext = "html"
    }

    h.dir, has = conf.String("dir")
    if !has {
        h.dir = fmt.Sprintf("template%v", os.PathSeparator)
    }

    h.funcs = template.FuncMap{
        "html": h.html,
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
        Logger.Fatalln("gmvc: missing template name")
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


func TemplateFuncs(funcs map[string]interface{}) {
    for k, f := range funcs {
        tpl.funcs[k] = f
    }
}


