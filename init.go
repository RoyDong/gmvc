package gmvc

import (
    "log"
    "os"
    "strings"
)

var (
    AppName  string
    Version  string
    Pwd      string
    Conf     *Tree
    ConfDir  = "config/"
    Env      = "prod"
    Host     = "0.0.0.0"
    Port     = 80
)


func initConfig() {
    confile := "config.json"
    for i, arg := range os.Args {
        if arg == "-c" && i+1 < len(os.Args) {
            confile = os.Args[i+1]
            if i := strings.LastIndex(confile, "/"); i >= 0 {
                Pwd = confile[:i+1]
            }
        }
    }

    Conf = NewTree()
    if err := Conf.LoadJson(confile, false); err != nil {
        log.Fatal("gmvc: ", err)
    }

    if v, ok := Conf.String("pwd"); ok {
        Pwd = v
    }
    if Pwd != "" {
        os.Chdir(Pwd)
    }

    if name, ok := Conf.String("name"); ok {
        AppName = name
    }

    if env, ok := Conf.String("env"); ok {
        Env = env
    }

    if v, ok := Conf.String("host"); ok {
        Host = v
    }

    if v, ok := Conf.Int("port"); ok {
        Port = v
    }

    if v, ok := Conf.String("template_ext"); ok {
        TemplateExt = v
    }

    if dir, ok := Conf.String("template_dir"); ok {
        if dir[len(dir)-1] != '/' {
            dir = dir + "/"
        }
        tpl.SetRootDir(dir)
    }
    if dir, ok := Conf.String("config_dir"); ok {
        if dir[len(dir)-1] != '/' {
            dir = dir + "/"
        }
        ConfDir = dir
    }
}

func Init() {
    event.Trigger("before_init")
    initConfig()
    event.Trigger("after_init")
}

