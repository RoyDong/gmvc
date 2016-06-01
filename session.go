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
    SessionDomain   string
    SessionDuration = 30
    SessionKeyName  = "GMVCSESSION"
)

type Session struct {
    key       string
    prefix    string
    updated   time.Time
    storage   *redis.Client
}

var sessions = make(map[string]*Session)

func NewSession(r *Request) *Session {
    conf := Conf.Tree("redis.main")

    s := &Session{
        prefix:     "session/",
        key:        genSessionKey(r.RemoteAddr),
        updated:    time.Now(),
        storage:    redis.DialTimeout(conf.Get("ip"), conf.Int64("port"), conf.Int64("timeout") * time.Second),
    }

    sessions[s.key] = s

    //set cookie
    http.SetCookie(r.rw, &http.Cookie{
        Name:     SessionKeyName,
        Value:    s.key,
        Path:     "/",
        Domain:   SessionDomain,
        HttpOnly: true,
    })

    return s
}

func (s *Session) SetPrefix(prefix string) {
    s.prefix = prefix
}

func (s *Session) RedisKey() string {
    return s.prefix + s.key + '/';
}

func (s *Session) Key() string {
    return s.key
}

func (s *Session) Set(key string, v interface{}) bool {
    return s.storage.Cmd("hset", s.RedisKey(), v)
}

func (s *Session) Get(key string) *redis.Resp {
    return s.storage.Cmd("hget", s.RedisKey())
}

func (s *Session) Clear() bool {
    return s.storage.Cmd("del", s.RedisKey())
}

/*
InitSession find session by session id set to request
if none found then create a new session
*/
func initSession(r *Request) {
    key := r.Header.Get(SessionKeyName)

    if key == nil {
        if cookie := r.Cookie(SessionKeyName); cookie != nil {
            key = cookie.Value
        }
    }

    if key != nil {
        var has bool
        if r.Session, has = sessions[key]; has {
            sec := time.Now().Unix()
            if has && r.Session.updated.Unix() + int64(SessionDuration) < sec {
                delete(sessions, r.Session.key)
                r.Session.Clear()
                r.Session = nil
                delete(sessions, key)
            }
        }
    }

    if r.Session == nil {
        r.Session = NewSession(r)
    }
}

func genSessionKey(salt string) string {
    rnd := make([]byte, 24)
    if _, err := io.ReadFull(rand.Reader, rnd); err != nil {
        panic(err.Error())
    }

    sig := fmt.Sprintf("%s%d%s", salt, time.Now().UnixNano(), rnd)
    hash := md5.New()
    if _, err := hash.Write([]byte(sig)); err != nil {
        panic(err.Error())
    }

    return hex.EncodeToString(hash.Sum(nil))
}

