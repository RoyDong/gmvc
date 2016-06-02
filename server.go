package gmvc

import (
    ws "github.com/gorilla/websocket"
    "fmt"
    "net"
    "net/http"
    "log"
    "sync"
    "strings"
)

/*
events:
    before_init
    after_init
    run

    request
    action
    respond
*/
var Event = NewEvent()

var route = &Route{}

func SetAction(action Action, patterns ...string) {
    for _, pattern := range patterns {
        route.Set(pattern, action)
    }
}

var ErrorAction = func(r *Request, c int, m string) *Response {
    resp := r.TextResponse(fmt.Sprintf("code: %d, message: %s", c, m))
    resp.SetStatus(c)
    return resp
}

type handler struct {
    ws.Upgrader
}

func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    req := newRequest(w, r)
    initSession(req)

    //websocket
    if strings.ToLower(r.Header.Get("Upgrade")) == "websocket" {
        conn, err := h.Upgrade(w, r, nil)
        if err != nil {
            panic(err.Error())
        }

        req.ws = conn
        defer conn.Close()
        Event.Trigger("ws_connect", req)
        defer Event.Trigger("ws_close", req)
        req.handleWSMessage()
        return
    }

    //normal request
    Event.Trigger("request", req)

    var act Action
    act, req.params = route.Parse(r.URL.Path)
    Event.Trigger("action", req)

    var resp *Response
    if act == nil {
        resp = ErrorAction(req, 404, "route not found")
    } else if resp = act(req); resp.code > 0 {
        resp = ErrorAction(req, resp.code, resp.message)
    }

    Event.Trigger("response", req, resp)
    if resp.status == 301 || resp.status == 302 {
        http.Redirect(w, r, resp.message, resp.status)
    } else {
        w.Write(resp.body)
    }
}

func listener() net.Listener {
    var err error
    var lsr net.Listener
    conf := Conf.Tree("server")
    lsr, err = net.Listen("tcp", fmt.Sprintf("%v:%v", conf.String("host"), conf.Int64("port")))
    if err != nil {
        panic(err.Error())
    }
    return lsr
}

var wg = &sync.WaitGroup{}

func serve() {
    defer wg.Done()
    srv := &http.Server{Handler: &handler{ws.Upgrader{ReadBufferSize: 1024, WriteBufferSize: 1024,}}}
    lsr := listener()
    defer lsr.Close()
    log.Println(srv.Serve(lsr))
}

func Run() {
    wg.Add(1)
    go serve()
    fmt.Println("work work")
    Event.Trigger("run")
    wg.Wait()
}
