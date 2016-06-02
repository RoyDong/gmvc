package gmvc

import (
    ws "github.com/gorilla/websocket"
    "encoding/json"
    "net/http"
    "net/url"
    "strconv"
    "strings"
    "bytes"
    "fmt"
    "log"
)

var tpl = newHtml()


func TemplateFuncs(funcs map[string]interface{}) {
    for k, f := range funcs {
        tpl.funcs[k] = f
    }
}

/*
consider it's the scope for an http request
*/
type Request struct {
    *http.Request
    Session *Session
    Cookies []*http.Cookie
    Bag     *Tree
    params  []string
    ws      *ws.Conn
    rw      http.ResponseWriter
}

func newRequest(w http.ResponseWriter, r *http.Request) *Request {
    return &Request{
        Request: r,
        Cookies: r.Cookies(),
        Bag: NewTree(),
        rw: w,
    }
}

func (r *Request) IsXMLHttpRequest() bool {
    return r.Header.Get("X-Requested-With") == "XMLHttpRequest"
}

func (r *Request) Int(k string) (int, bool) {
    if v, has := r.String(k); has {
        if i, err := strconv.ParseInt(v, 10, 0); err == nil {
            return int(i), true
        }
    }
    return 0, false
}

func (r *Request) Int64(k string) (int64, bool) {
    if v, has := r.String(k); has {
        if i, err := strconv.ParseInt(v, 10, 64); err == nil {
            return i, true
        }
    }
    return 0, false
}

func (r *Request) Float(k string) (float64, bool) {
    if v, has := r.String(k); has {
        if f, err := strconv.ParseFloat(v, 64); err == nil {
            return f, true
        }
    }
    return 0, false
}

func (r *Request) String(k string) (string, bool) {
    if k[0] == '$' {
        n, err := strconv.ParseInt(k[1:], 10, 0)
        if err == nil && n > 0 && int(n) <= len(r.params) {
            return r.params[n-1], true
        }
    }
    if v := r.FormValue(k); len(v) > 0 {
        return v, true
    }
    return "", false
}

func (r *Request) Cookie(name string) *http.Cookie {
    for _, c := range r.Cookies {
        if c.Name == name {
            return c
        }
    }
    return nil
}


/*
websocket action
*/
type WSAction func(wsm *WSMessage)

var WSActionMap = make(map[string]WSAction)


/*
websocket message
*/
type WSMessage struct {
    Name    string
    Query   map[string]string
    Data    []byte
    Request *Request
    Bag     *Tree
}


func (r *Request) newWSMessage(raw []byte) *WSMessage {
    parts := bytes.Split(raw, bytes.TrimLeft(raw, "\n"))
    wsm := &WSMessage{Name: string(parts[0]), Request: r, Bag: NewTree()}
    if len(parts[1]) > 0 {
        values, err := url.ParseQuery(string(parts[1]))
        if err == nil {
            wsm.Query = make(map[string]string, len(values))
            for k, vs := range values {
                if len(vs) > 0 {
                    wsm.Query[k] = vs[0]
                }
            }
        }
    }
    if len(parts[2]) > 0 {
        wsm.Data = parts[2]
    }
    return wsm
}

func (r *Request) SendWSMessage(name string, query map[string]interface{}, data interface{}) {
    q := make([]string, 0, len(query))
    for k, v := range query {
        q = append(q, fmt.Sprintf("%s=%v", k, v))
    }

    json, err := json.Marshal(data)
    if err != nil {
        log.Println(err.Error());
    }

    r.ws.WriteMessage(ws.TextMessage, []byte(name + "\n" + strings.Join(q, "&") + "\n" + string(json) + "\n"))
}

func (wsm *WSMessage) Send(name string, query map[string]interface{}, data interface{}) {
    wsm.Request.SendWSMessage(name, query, data)
}

func (wsm *WSMessage) Decode(v interface{}) {
    if err := json.Unmarshal(wsm.Data, v); err != nil {
        log.Println(err.Error())
    }
}

func (wsm *WSMessage) String(key string) (string, bool) {
    if len(wsm.Query) > 0 {
        v, has := wsm.Query[key]
        return v, has
    }
    return "", false
}

func (wsm *WSMessage) Int(k string) (int, bool) {
    if v, has := wsm.String(k); has {
        if i, err := strconv.ParseInt(v, 10, 0); err == nil {
            return int(i), true
        }
    }
    return 0, false
}

func (wsm *WSMessage) Int64(k string) (int64, bool) {
    if v, has := wsm.String(k); has {
        if i, err := strconv.ParseInt(v, 10, 64); err == nil {
            return i, true
        }
    }
    return 0, false
}

func (wsm *WSMessage) Float(k string) (float64, bool) {
    if v, has := wsm.String(k); has {
        if f, err := strconv.ParseFloat(v, 64); err == nil {
            return f, true
        }
    }
    return 0, false
}

func (r *Request) handleWSMessage() {
    if r.ws == nil {
        panic("gmvc: bad websocket connection")
    }
    for {
        var raw []byte
        var err error
        if _, raw, err = r.ws.ReadMessage(); err != nil {
            log.Println(err)
            return
        }
        go func() {
            defer func() {
                if err := recover(); err != nil {
                    log.Println("gmvc: websocket ", err)
                }
            }()
            var wsm = r.newWSMessage(raw)
            if wsa, has := WSActionMap[wsm.Name]; has {
                wsa(wsm)
            } else {
                log.Println("gmvc: wsa " + wsm.Name + "not found")
            }
        }()
    }
}


type Response struct {
    status  int
    code    int
    message string
    body    []byte
    rw      http.ResponseWriter
}

func (r *Request) newResponse() *Response {
    return &Response{rw: r.rw}
}

func (r *Request) TextResponse(txt string) *Response {
    p := r.newResponse()
    p.body = []byte(txt)
    return p
}

func (r *Request) HtmlResponse(data interface{}, args ...string) *Response {
    resp := r.newResponse()
    resp.body = tpl.render(data, args...)
    return resp
}

func (r *Request) JsonResponse(data interface{}) *Response {
    json, err := json.Marshal(data)
    if err != nil {
        //TODO 错误处理
        log.Println(err.Error());
    }

    p := r.newResponse()
    p.rw.Header().Set("Content-Type", "application/json;")
    p.body = json
    return p
}

func (r *Request) ErrorResponse(code int, msg string) *Response {
    p := r.newResponse()
    p.code = code
    p.message = msg
    return p
}

func (r *Request) RedirectResponse(url string, status int) *Response {
    p := r.newResponse()
    p.status = status
    p.message = url
    return p
}

func (p *Response) Header() http.Header {
    return p.rw.Header()
}

func (p *Response) SetStatus(status int) {
    p.rw.WriteHeader(status)
}

func (p *Response) SetCookie(c *http.Cookie) {
    http.SetCookie(p.rw, c)
}

func (p *Response) Body() []byte {
    return p.body
}


