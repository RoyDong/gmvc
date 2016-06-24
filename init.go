package gmvc

import (
    "os"
    "strings"
)

var (
    Pwd      string
    Store    *Tree
    Env      = "prod"

    tpl      *html
)


func initStore() {
    confdir := "config"
    for i, arg := range os.Args {
        if arg == "-c" && i+1 < len(os.Args) {
            confdir = os.Args[i+1]
        }
    }

    fd, e := os.Open(confdir)
    if e != nil {
        panic("gmvc: config dir not found")
    }
    defer fd.Close()

    dinfo, e := fd.Readdir(-1)
    if e != nil {
        panic("gmvc: read config dir error")
    }

    hasConfig := false
    Store = NewTree()
    for _, info := range dinfo {
        if !info.IsDir() && strings.HasSuffix(info.Name(), ".yml") {
            key := strings.TrimRight(info.Name(), ".yml")
            Store.LoadYamlFile(key, fd.Name() + "/" + info.Name(), false)
            if key == "config" {
                hasConfig = true
            }
        }
    }

    if !hasConfig {
        panic("gmvc: must have a config.yml")
    }

    if v, ok := Store.String("config.pwd"); ok {
        Pwd = v
        os.Chdir(Pwd)
    }

    if env, ok := Store.String("config.env"); ok {
        Env = env
    }
}

func init() {
    Hook.Trigger("before_init")

    initStore()

    Logger = initLogger("info")
    if Logger == nil {
        panic("gmvc: init logger info")
    }

    accesslog = initLogger("access")
    if accesslog == nil {
        panic("gmvc: init logger error")
    }

    tpl = newHtml()

    initSession()

    Hook.Trigger("after_init")
}


