package gmvc

import (
    "fmt"
    "time"
    "github.com/mediocregopher/radix.v2/redis"
    "sync"
)


var (
    redisClients = make(map[string]*redis.Client)
    redisClientsLocker = &sync.Mutex{}
)

func RedisClient(name string) *redis.Client {
    client, has := redisClients[name]
    if !has {
        client = newRedisClient(name)
        redisClients[name] = client
    }

    return client
}


func newRedisClient(name string) *redis.Client {
    redisClientsLocker.Lock()
    defer redisClientsLocker.Unlock()

    if client, has := redisClients[name]; has {
        return client
    }

    conf := Store.Tree("config.redis." + name)
    ip, _ := conf.String("ip")
    port, _ := conf.Int64("port")
    timeout, _ := conf.Int("timeout")

    client, err := redis.DialTimeout("tcp", fmt.Sprintf("%s:%d", ip, port), time.Duration(timeout) * time.Second);
    if err != nil {
        Logger.Fatalln("gmvc: can not connect to redis " + name)
    }

    return client
}
