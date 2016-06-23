package gmvc

import (
    "net/http"
    "time"
    "github.com/mediocregopher/radix.v2/redis"
    "fmt"
)

var (
    SessionDomain   = ""
    SessionExpires  = int64(1800)
    SessionKeyName  = "GMVCSESSION"
    SessionPrefix   = "session/"
    SessionEnabled  = false
    SessionRedis    = "main"


    sessions        = make(map[string]*Session)
    sessionStore    *redis.Client
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
    resp := sessionStore.Cmd("hadd", s.RedisKey(), v)
    ret, err := resp.Int64()
    if err != nil {
        return false
    }
    return ret > 0
}

func (s *Session) Set(key string, v interface{}) bool {
    resp := sessionStore.Cmd("hset", s.RedisKey(), key, v)
    ret, err := resp.Int64()
    if err != nil {
        return false
    }
    return ret > 0
}

func (s *Session) Get(key string) *redis.Resp {
    return sessionStore.Cmd("hget", s.RedisKey(), key)
}

func (s *Session) Clear() {
    sessionStore.Cmd("del", s.RedisKey())
}


/*
InitSession find session by session id set to request
if none found then create a new session
*/
func initSession() {
    conf := Store.Tree("config.session")

    if v, ok := conf.Int("enabled"); ok {
        if v > 0 {
            SessionEnabled = true
        } else {
            SessionEnabled = false
        }
    }

    if !SessionEnabled {
        return
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

    if v ,ok := conf.String("redis"); ok {
        SessionRedis = v
    }

    sessionStore = RedisClient(SessionRedis)
}

func retrieveSession(r *Request) {
    if !SessionEnabled {
        return
    }

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
            sessionStore.Cmd("del", s.RedisKey())
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
    return MD5(fmt.Sprintf("%s%d%s", salt, time.Now().UnixNano(), RandString(24)))
}

