package gmvc

import (
    "regexp"
    "strings"
    "log"
    "fmt"
)

type Action func(r *Request) *Response

type Router struct {
    name   string
    regexp *regexp.Regexp
    action Action

    statics map[string]*Router
    regexps []*Router
}

func (r *Router) Parse(path string) (Action, []string) {
    current := r
    params := make([]string, 0)
    nodes := strings.Split(strings.ToLower(strings.Trim(path, "/")), "/")
    for _, name := range nodes {
        var found bool
        var router *Router
        if len(current.statics) > 0 {
            if router, found = current.statics[name]; found {
                current = router
            }
        }
        if !found {
            for _, route := range current.regexps {
                subs := route.regexp.FindStringSubmatch(name)
                if len(subs) >= 2 {
                    params = append(params, subs[1:]...)
                    current = route
                    found = true
                    break
                }
            }
        }
        if !found {
            return nil, nil
        }
    }
    return current.action, params
}

func (r *Router) Set(path string, action Action) {
    current := r
    nodes := strings.Split(strings.ToLower(strings.Trim(path, "/")), "/")
    for _, name := range nodes {
        rt, found := current.statics[name]
        if !found {
            for _, rt = range current.regexps {
                if name == rt.name {
                    current = rt
                    found = true
                    break
                }
            }
        }
        if !found {
            rt = &Router{name: name}
            if strings.Contains(name, "(") {
                rt.regexp = regexp.MustCompile("^" + name + "$")
                if current.regexps == nil {
                    current.regexps = make([]*Router, 0, 1)
                }
                current.regexps = append(current.regexps, rt)
            } else {
                if current.statics == nil {
                    current.statics = make(map[string]*Router, 1)
                }
                current.statics[name] = rt
            }
        }
        current = rt
    }
    if current.action != nil {
        log.Println(fmt.Sprintf("gmvc: %s already had an action and it will be overwritten", path))
    }
    current.action = action
}
