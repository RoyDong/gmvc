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
    "crypto/sha512"
)

var (
    SessionDomain   = ""
    SessionExpires  = int64(1800)
    SessionKeyName  = "GMVCSESSION"
    SessionPrefix   = "session/"

    sessions        = make(map[string]*Session)
    sessionStorage  *redis.Client
)

type Session struct {
    key       string
    updated   time.Time
}

func (s *Session) RedisKey() string {
    return SessionPrefix + s.key + "/";
}

func (s *Session) Key() string {
    return s.key
}

func (s *Session) Add(key string, v interface{}) bool {
    resp := sessionStorage.Cmd("hadd", s.RedisKey(), v)
    ret, err := resp.Int64()
    if err != nil {
        return false
    }
    return ret > 0
}

func (s *Session) Set(key string, v interface{}) bool {
    resp := sessionStorage.Cmd("hset", s.RedisKey(), key, v)
    ret, err := resp.Int64()
    if err != nil {
        return false
    }
    return ret > 0
}

func (s *Session) Get(key string) *redis.Resp {
    return sessionStorage.Cmd("hget", s.RedisKey(), key)
}

func (s *Session) Clear() {
    sessionStorage.Cmd("del", s.RedisKey())
}


/*
InitSession find session by session id set to request
if none found then create a new session
*/
func initSession() {
    conf := Conf.Tree("session")

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

    rconf := conf.Tree("redis")
    ip, _ := rconf.String("ip")
    port, _ := rconf.Int64("port")
    timeout, _ := rconf.Int("timeout")

    var err error
    sessionStorage, err = redis.DialTimeout("tcp", fmt.Sprintf("%s:%d", ip, port), time.Duration(timeout) * time.Second);
    if err != nil {
        Logger.Fatalln(err.Error())
    }
}

func retrieveSession(r *Request) {
    key := r.Header.Get(SessionKeyName)

    if len(key) == 0 {
        if cookie := r.Cookie(SessionKeyName); cookie != nil {
            key = cookie.Value
        }
    }

    var s *Session
    var has bool
    if s, has = sessions[key]; !has {
        s = &Session{key: key}
    }

    if len(s.key) == 0 {
        s.key = genSessionKey(r.RemoteAddr)
    } else {
        sec, _ := s.Get("__updated").Int64()
        now := time.Now().Unix()
        if sec + SessionExpires < now {
            sessionStorage.Cmd("del", s.RedisKey())
            delete(sessions, s.key)
            s.key = genSessionKey(r.RemoteAddr)
        }
    }

    s.updated = time.Now()
    s.Set("__updated", s.updated.Unix())

    sessions[s.key] = s
    r.Session = s

    //set cookie
    http.SetCookie(r.rw, &http.Cookie{
        Name:     SessionKeyName,
        Value:    s.key,
        Path:     "/",
        Domain:   SessionDomain,
        HttpOnly: true,
    })
}

func genSessionKey(salt string) string {
    rnd := make([]byte, 24)
    if _, err := io.ReadFull(rand.Reader, rnd); err != nil {
        Logger.Fatalln(err.Error())
    }

    sig := fmt.Sprintf("%s%d%s", salt, time.Now().UnixNano(), rnd)
    hash := md5.New()
    if _, err := hash.Write([]byte(sig)); err != nil {
        Logger.Fatalln(err.Error())
    }

    return hex.EncodeToString(hash.Sum(nil))
}

