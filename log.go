package gmvc

import (
    "log"
    "os"
)

var (
    Logger    *log.Logger
    accesslog *log.Logger

    DirPerm   = 0755
    LogPerm   = os.FileMode(0644)
)


func initLogger(name string) *log.Logger {
    conf := Conf.Tree("log")

    file, has := conf.String(name)
    if !has {
        file = "log/" + name + ".log"
    }

    out, err := createLogfile(file)
    if err == nil {
        return log.New(out, "gmvc", 1)
    }
    return nil
}

func createLogfile(filename string) (*os.File, error) {
    var f *os.File
    var e error
    f, e = os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, LogPerm)
    if e != nil {
        return nil, e
    }
    return f, nil
}

