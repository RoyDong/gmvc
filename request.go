package gmvc

import (
    "github.com/gorilla/websocket"
    "encoding/json"
    "html/template"
    "net/http"
    "net/url"
    "strconv"
    "strings"
    "bytes"
    "fmt"
    "log"
)

const (
    StatusBadRequest              = 400
    StatusPaymentRequired         = 402
    StatusForbidden               = 403
    StatusNotFound                = 404
    StatusUnsupportedMediaType    = 415
    StatusInternalServerError     = 500
    StatusServiceUnavailable      = 503
    StatusGatewayTimeout          = 504
    StatusHTTPVersionNotSupported = 505
)

var tpl = NewTemplate()

func TemplateFuncs(funcs map[string]interface{}) {
    tpl.AddFuncs(funcs)
}

/*
consider it's the scope for an http request
*/
type Request struct {
    *http.Request
    Session *Session
    Cookies []*http.Cookie
    Bag     *lib.Tree
    params  []string
    ws      *ws.Conn
    rw      http.ResponseWriter
}

func newRequest(w http.ResponseWriter, r *http.Request) *Request {
    return &Request{
        Request: r,
        Cookies: r.Cookies(),
        Bag: lib.NewTree(),
        rw: w,
    }
}

func (r *Request) IsAjax() bool {
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
type Wsa func(wsm *Wsm)

var WsaMap = make(map[string]Wsa)

/*
websocket message
*/
type Wsm struct {
    Name    string
    Query   map[string]string
    Data    []byte
    Request *Request          `codec:"-"`
    Bag     *lib.Tree         `codec:"-"`
}


var msgpackHandle = new(codec.MsgpackHandle)

func (r *Request) newWsm(raw []byte) *Wsm {
    parts := bytes.Split(raw, bytes.TrimLeft(raw, "\n"))
    wsm := &Wsm{Name: string(parts[0]), Request: r, Bag: lib.NewTree()}
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

func (r *Request) SendWsm(name string, query map[string]interface{}, data interface{}) {
    q := make([]string, 0, len(query))
    for k, v := range query {
        q = append(q, fmt.Sprintf("%s=%v", k, v))
    }
    var d []byte
    if data != nil {
        var enc *codec.Encoder
        enc = codec.NewEncoderBytes(&d, msgpackHandle)
        if err := enc.Encode(data); err != nil {
            panic("gmvc: " + err.Error())
        }
    }
    ws.Message.Send(r.ws, name + "\n" + strings.Join(q, "&") + "\n" + string(d) + "\n")
}

func (wsm *Wsm) Send(name string, query map[string]interface{}, data interface{}) {
    wsm.Request.SendWsm(name, query, data)
}

func (wsm *Wsm) Decode(v interface{}) {
    dec := codec.NewDecoderBytes(wsm.Data, msgpackHandle)
    if err := dec.Decode(v); err != nil {
        panic("gmvc: " + err.Error())
    }
}

func (wsm *Wsm) String(key string) (string, bool) {
    if len(wsm.Query) > 0 {
        v, has := wsm.Query[key]
        return v, has
    }
    return "", false
}

func (wsm *Wsm) Int(k string) (int, bool) {
    if v, has := wsm.String(k); has {
        if i, err := strconv.ParseInt(v, 10, 0); err == nil {
            return int(i), true
        }
    }
    return 0, false
}

func (wsm *Wsm) Int64(k string) (int64, bool) {
    if v, has := wsm.String(k); has {
        if i, err := strconv.ParseInt(v, 10, 64); err == nil {
            return i, true
        }
    }
    return 0, false
}

func (wsm *Wsm) Float(k string) (float64, bool) {
    if v, has := wsm.String(k); has {
        if f, err := strconv.ParseFloat(v, 64); err == nil {
            return f, true
        }
    }
    return 0, false
}

func (r *Request) handleWs() {
    if r.ws == nil {
        panic("gmvc: normal request no websocket")
    }
    for {
        var raw []byte
        if err := ws.Message.Receive(r.ws, &raw); err != nil {
            log.Println(err)
            return
        }
        go func() {
            defer func() {
                if err := recover(); err != nil {
                    log.Println("gmvc: websocket ", err)
                }
            }()
            var wsm = r.newWsm(raw)
            if wsa, has := WsaMap[wsm.Name]; has {
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

func (r *Request) HtmlResponse(name string, data interface{}) *Response {
    var t *template.Template
    buffer := &bytes.Buffer{}
    resp := r.newResponse()
    html := NewHtml()
    html.Data = data
    t = tpl.Template(name)
    if t == nil {
        panic("gmvc: " + name + " template not found")
    }
    t.Execute(buffer, html)

    //has layout
    if html.layout != "" {
        t = tpl.Template(html.layout)
        if t == nil {
            panic("gmvc: " + html.layout + " template not found")
        }
        html.Content = template.HTML(buffer.Bytes())
        buffer.Truncate(0)
        t.Execute(buffer, html)
    }
    resp.body = buffer.Bytes()
    return resp
}

func (r *Request) JsonResponse(data interface{}) *Response {
    json, err := json.Marshal(data)
    if err != nil {
        panic("gmvc: " + err.Error())
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
