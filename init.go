package gmvc

import (
    "log"
    "os"
    "strings"
)

var (
    Pwd      string
    Conf     *Tree
    Env      = "prod"
)


func initConfig() {
    confile := "config.yml"
    for i, arg := range os.Args {
        if arg == "-c" && i+1 < len(os.Args) {
            confile = os.Args[i+1]
            if i := strings.LastIndex(confile, "/"); i >= 0 {
                Pwd = confile[:i+1]
            }
        }
    }

    Conf = NewTree()
    if err := Conf.LoadYaml(confile, false); err != nil {
        log.Fatal("gmvc: ", err)
    }

    if v, ok := Conf.String("pwd"); ok {
        Pwd = v
    }
    if Pwd != "" {
        os.Chdir(Pwd)
    }

    if env, ok := Conf.String("env"); ok {
        Env = env
    }

    if v, ok := Conf.String("template.ext"); ok {
        tpl.ext = v
    }

    if dir, ok := Conf.String("template.dir"); ok {
        if dir[len(dir)-1] != '/' {
            dir = dir + "/"
        }
        tpl.dir = dir
    }
}

func Init() {
    Hook.Trigger("before_init")
    initConfig()
    Hook.Trigger("after_init")
}

