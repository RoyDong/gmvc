package gmvc

import (
    "github.com/gorilla/websocket"
    "fmt"
    "net"
    "net/http"
    "log"
    "os"
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
var event = NewEvent()

func AddHandler(name string, handler EventHandler) {
    event.AddHandler(name, handler)
}

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
    ws.Server
}

func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    req := newRequest(w, r)
    initSession(req)

    //websocket
    if strings.ToLower(r.Header.Get("Upgrade")) == "websocket" {
        if conn := h.Conn(w, r); conn != nil {
            req.ws = conn
            defer conn.Close()
            event.Trigger("ws_connect", req)
            defer event.Trigger("ws_close", req)
            req.handleWs()
        }
        return
    }

    //normal request
    event.Trigger("request", req)
    if Env == "dev" {
        tpl.Load(TplDir)
    }

    var act Action
    act, req.params = route.Parse(r.URL.Path)
    event.Trigger("action", req)

    var resp *Response
    if act == nil {
        resp = ErrorAction(req, 404, "route not found")
    } else if resp = act(req); resp.code > 0 {
        resp = ErrorAction(req, resp.code, resp.message)
    }

    event.Trigger("respond", req, resp)
    if resp.status == 301 || resp.status == 302 {
        http.Redirect(w, r, resp.message, resp.status)
    } else {
        w.Write(resp.body)
    }
}

func listener() net.Listener {
    var err error
    var lsr net.Listener
    if len(SockFile) > 0 {
        os.Remove(SockFile)
        lsr, err = net.Listen("unix", SockFile)
        if err == nil {
            os.Chmod(SockFile, os.ModePerm)
            return lsr
        }
    }
    lsr, err = net.Listen("tcp", fmt.Sprintf(":%d", Port))
    if err != nil {
        log.Fatal("gmvc:", err)
    }
    return lsr
}

var wg = &sync.WaitGroup{}

func serve() {
    defer wg.Done()
    srv := &http.Server{Handler: &handler{ws.Server{}}}
    lsr := listener()
    defer lsr.Close()
    log.Println(srv.Serve(lsr))
}

func Run() {
    tpl.Load(TplDir)
    go sessionExpire()
    wg.Add(1)
    go serve()
    fmt.Println("work work")
    event.Trigger("run")
    wg.Wait()
}
