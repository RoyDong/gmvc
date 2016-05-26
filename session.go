package gmvc

import (
    "crypto/md5"
    "crypto/rand"
    "encoding/hex"
    "fmt"
    "io"
    "net/http"
    "time"
    "github.com/mediocregopher/radix.v2/redis"
)

var (
    SessionDomain     string
    SessionDuration   = 30
    SessionCookieName = "GMVCSESSION"
)

type Session struct {
    *Tree
    Id        string
    UpdatedAt time.Time
}

var sessions = make(map[string]*Session)

func NewSession(r *Request) *Session {
    s := &Session{
        Tree:      NewTree(),
        Id:        sessionId(r),
        UpdatedAt: time.Now(),
    }
    //set cookie
    sessions[s.Id] = s
    http.SetCookie(r.rw, &http.Cookie{
        Name:     SessionCookieName,
        Value:    s.Id,
        Path:     "/",
        Domain:   SessionDomain,
        HttpOnly: true,
    })
    return s
}

/*
InitSession find session by session id set to request
if none found then create a new session
*/
func initSession(r *Request) {
    if c := r.Cookie(SessionCookieName); c != nil {
        var has bool
        if r.Session, has = sessions[c.Value]; has {
            sec := time.Now().Unix()
            if has && r.Session.UpdatedAt.Unix() + int64(SessionDuration) < sec {
                delete(sessions, r.Session.Id)
                r.Session.Clear()
                r.Session = nil
            }
        }
    }
    if r.Session == nil {
        r.Session = NewSession(r)
    }
}

func sessionId(r *Request) string {
    rnd := make([]byte, 24)
    if _, err := io.ReadFull(rand.Reader, rnd); err != nil {
        panic("gmvc: session id " + err.Error())
    }

    sig := fmt.Sprintf("%s%d%s", r.RemoteAddr, time.Now().UnixNano(), rnd)
    hash := md5.New()
    if _, err := hash.Write([]byte(sig)); err != nil {
        panic("gmvc: session id " + err.Error())
    }

    return hex.EncodeToString(hash.Sum(nil))
}

/**
sessionExpire checks sessons expiration per minute
and delete all expired sessions
*/
func sessionExpire() {
    for now := range time.Tick(time.Minute) {
        t := now.Unix()
        for k, s := range sessions {
            if s.UpdatedAt.Unix() + int64(SessionDuration) < t {
                s.Clear()
                delete(sessions, k)
            }
        }
    }
}
