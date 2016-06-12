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
    confile := "config.yml"
    for i, arg := range os.Args {
        if arg == "-c" && i+1 < len(os.Args) {
            confile = os.Args[i+1]
            if i := strings.LastIndex(confile, "/"); i >= 0 {
                Pwd = confile[:i+1]
            }
        }
    }

    Store = NewTree()
    if err := Store.LoadYaml("config", confile, false); err != nil {
        panic(err)
    }

    if v, ok := Store.String("config.pwd"); ok {
        Pwd = v
    }
    if Pwd != "" {
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
    accesslog = initLogger("access")

    tpl = newHtml()

    initSession()

    Hook.Trigger("after_init")
}


