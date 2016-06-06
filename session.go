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
    SessionDomain   = ""
    SessionExpires  = int64(1800)
    SessionKeyName  = "GMVCSESSION"
    SessionPrefix   = "session/"
)

type Session struct {
    key       string
    updated   time.Time
    storage   *redis.Client
}

var sessions = make(map[string]*Session)

func newSession(r *Request) *Session {
    conf := Conf.Tree("session")
    storage := conf.Tree("redis")

    ip, _ := storage.String("ip")
    port, _ := storage.Int64("port")
    timeout, _ := storage.Int("timeout")

    client, err := redis.DialTimeout("tcp", fmt.Sprintf("%s:%d", ip, port), time.Duration(timeout) * time.Second);
    if err != nil {
        panic(err.Error())
    }

    s := &Session{
        key:        genSessionKey(r.RemoteAddr),
        updated:    time.Now(),
        storage:    client,
    }

    if v, ok := conf.String("prefix"); ok {
        SessionPrefix = v
    }

    if v, ok := conf.String("keyname"); ok {
        SessionKeyName = v
    }

    if v, ok := conf.String("domain"); ok {
        SessionDomain = v
    }

    if v, ok := conf.Int64("expires"); ok {
        SessionExpires = v
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

func (s *Session) RedisKey() string {
    return SessionPrefix + s.key + "/";
}

func (s *Session) Key() string {
    return s.key
}

func (s *Session) Add(key string, v interface{}) bool {
    resp := s.storage.Cmd("hadd", s.RedisKey(), v)
    ret, err := resp.Int64()
    if err != nil {
        return false
    }
    return ret > 0
}

func (s *Session) Set(key string, v interface{}) bool {
    resp := s.storage.Cmd("hset", s.RedisKey(), v)
    ret, err := resp.Int64()
    if err != nil {
        return false
    }
    return ret > 0
}

func (s *Session) Get(key string) *redis.Resp {
    return s.storage.Cmd("hget", s.RedisKey())
}

func (s *Session) Clear() {
    s.storage.Cmd("del", s.RedisKey())
}

/*
InitSession find session by session id set to request
if none found then create a new session
*/
func initSession(r *Request) {
    key := r.Header.Get(SessionKeyName)

    if len(key) == 0 {
        if cookie := r.Cookie(SessionKeyName); cookie != nil {
            key = cookie.Value
        }
    }

    if len(key) > 0 {
        var has bool
        if r.Session, has = sessions[key]; has {
            sec := time.Now().Unix()
            if has && r.Session.updated.Unix() + SessionExpires < sec {
                delete(sessions, r.Session.key)
                r.Session.Clear()
                r.Session = nil
            }
        }
    }

    if r.Session == nil {
        r.Session = newSession(r)
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

