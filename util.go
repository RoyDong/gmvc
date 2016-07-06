package gmvc

import (
    "math"
    "github.com/gorilla/websocket"
    "sync"
)

func Round(f float64, n int) float64 {
    base := math.Pow10(n)

    f = f * base

    if f < 0 {
        f = math.Ceil(f - 0.5) / base
    } else {
        f = math.Floor(f + 0.5) / base
    }

    return f
}

func Max(x, y int64) int64 {
    if x >= y {
        return x
    }
    return y
}

var (
    wsconns = make(map[string]*websocket.Conn)
    wsconnsLocker = &sync.Mutex{}
)


func WSConn(uri string) *websocket.Conn {
    conn, has := wsconns[uri]
    if !has {
        conn = newWSConn(uri)
        if conn != nil {
            wsconns[uri] = conn
        }
    }
    return conn
}

func newWSConn(url string) *websocket.Conn {
    wsconnsLocker.Lock()
    defer wsconnsLocker.Unlock()

    if conn, has := wsconns[url]; has {
        return conn
    }

    dialer := &websocket.Dialer{
        ReadBufferSize: 1024,
        WriteBufferSize: 1024,
    }
    conn, _, err := dialer.Dial(url, nil)
    if err != nil {
        Logger.Println(err)
        return nil
    }
    return conn
}



