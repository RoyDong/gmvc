package gmvc

import (
    "os"
)

var (
    Pwd      string
    Store    *Tree
    Env      = "prod"

    tpl      *html
)


func initStore() {
    confile := "config.yml"
    for i, arg := range os.Args {
        if arg == "-c" && i+1 < len(os.Args) {
            confile = os.Args[i+1]
        }
    }

    Store = NewTree()
    if err := Store.LoadYaml("config", confile, false); err != nil {
        panic("gmvc: config file is missing")
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

    Logger = initLogger("error")
    if Logger == nil {
        panic("gmvc: init logger error")
    }

    accesslog = initLogger("access")
    if accesslog == nil {
        panic("gmvc: init logger error")
    }

    tpl = newHtml()

    initSession()

    Hook.Trigger("after_init")
}


