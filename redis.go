package gmvc

import (
    "github.com/mediocregopher/radix.v2/redis"
    "fmt"
    "time"
)


type Redis struct {
    client *redis.Client
}

func NewRedis(host string, port int, timeout time.Duration) *Redis {
    client, err := redis.DialTimeout("tcp", fmt.Sprintf("%v:%v", host, port), timeout)
    if (err != nil) {
        panic(err.Error())
    }

    return &Redis{
        client: client,
    }
}

func (r *Redis) Cmd(name string, args ...interface{}) *redis.Resp {
    return r.client.Cmd(name, args...)
}

