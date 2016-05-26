example: https://github.com/RoyDong/notes

modify go.net/websocket/server.go add a function:

    func (s Server) Conn(w http.ResponseWriter, req *http.Request) *Conn {
        rwc, buf, err := w.(http.Hijacker).Hijack()
        if err != nil {
            return nil
        }
        conn, err := newServerConn(rwc, buf, req, &s.Config, s.Handshake)
        if err != nil {
            return nil
        }
        if conn == nil {
            return nil
        }
        return conn
    }


