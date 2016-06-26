package gmvc

import (
    ws "github.com/gorilla/websocket"
    "fmt"
    "net"
    "net/http"
    "sync"
    "strings"
    "os"
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
var Hook = NewEvent()

var router = &Router{}

func SetAction(action Action, patterns ...string) {
    for _, pattern := range patterns {
        router.Set(pattern, action)
    }
}

var ErrorAction = func(r *Request, c int, m string) *Response {
    resp := r.TextResponse(fmt.Sprintf("code: %d, message: %s", c, m))
    resp.SetStatus(c)
    return resp
}

type handler struct {
    ws.Upgrader
    sdir string
    sprefix string
}

func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    req := newRequest(w, r)
    retrieveSession(req)

    path := r.URL.Path
    accesslog.Println(path)

    //websocket
    if strings.ToLower(r.Header.Get("Upgrade")) == "websocket" {
        conn, err := h.Upgrade(w, r, nil)
        if err != nil {
            Logger.Fatalln(err.Error())
        }

        req.ws = conn
        defer conn.Close()
        Hook.Trigger("ws_connect", req)
        defer Hook.Trigger("ws_close", req)
        req.handleWSMessage()
        return
    }

    //static files
    if strings.HasPrefix(path, h.sprefix) {
        filename := Pwd + h.sdir + strings.TrimLeft(path, h.sprefix)
        http.ServeFile(w, r, filename)
        return
    }

    //normal request
    Hook.Trigger("request", req)

    var act Action
    act, req.params = router.Parse(path)
    Hook.Trigger("action", req)

    var resp *Response
    if act == nil {
        resp = ErrorAction(req, 404, "route not found")
    } else if resp = act(req); resp.code > 0 {
        resp = ErrorAction(req, resp.code, resp.message)
    }

    Hook.Trigger("response", req, resp)
    if resp.status == 301 || resp.status == 302 {
        http.Redirect(w, r, resp.message, resp.status)
    } else {
        w.Write(resp.body)
    }
}

var wg = &sync.WaitGroup{}

func serve() {
    defer wg.Done()

    conf := Store.Tree("config.server")
    host, _ := conf.String("host")
    port, _ := conf.Int64("port")
    lsr, err := net.Listen("tcp", fmt.Sprintf("%v:%v", host, port))
    defer lsr.Close()
    if err != nil {
        Logger.Fatalln(err.Error())
    }

    Logger.Println(fmt.Sprintf("listening %v:%v", host, port))

    h := &handler{
        Upgrader: ws.Upgrader{ReadBufferSize: 1024, WriteBufferSize: 1024},
        sdir: fmt.Sprintf("static%v", os.PathSeparator),
        sprefix: "/static",
    }

    if v, has := Store.String("config.static_file.dir"); has {
        h.sdir = v
    }

    if v, has := Store.String("config.static_file.prefix"); has {
        h.sprefix = v
    }

    srv := &http.Server{Handler: h}
    Logger.Println(srv.Serve(lsr))
}

func Run() {
    wg.Add(1)
    go serve()
    Hook.Trigger("run")
    wg.Wait()
}

