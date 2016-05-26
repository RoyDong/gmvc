package gmvc

import (
    "github.com/mediocregopher/radix.v2/redis"
    "github.com/ugorji/go/codec"
)


type Redis struct {
    client redis.Client
}



