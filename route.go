package gmvc

import (
    "regexp"
    "strings"
)

type Action func(r *Request) *Response

type Route struct {
    name   string
    regexp *regexp.Regexp
    action Action

    statics map[string]*Route
    regexps []*Route
}

func (r *Route) Parse(path string) (Action, []string) {
    current := r
    params := make([]string, 0)
    nodes := strings.Split(strings.ToLower(strings.Trim(path, "/")), "/")
    for _, name := range nodes {
        var found bool
        var route *Route
        if len(current.statics) > 0 {
            if route, found = current.statics[name]; found {
                current = route
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

func (r *Route) Set(path string, action Action) {
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
            rt = &Route{name: name}
            if strings.Contains(name, "(") {
                rt.regexp = regexp.MustCompile("^" + name + "$")
                if current.regexps == nil {
                    current.regexps = make([]*Route, 0, 1)
                }
                current.regexps = append(current.regexps, rt)
            } else {
                if current.statics == nil {
                    current.statics = make(map[string]*Route, 1)
                }
                current.statics[name] = rt
            }
        }
        current = rt
    }
    current.action = action
}
