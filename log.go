package gmvc

import (
    "log"
    "os"
)

var (
    Logger    *log.Logger
    accesslog *log.Logger
)


func initLogger(name string) *log.Logger {
    conf := Store.Tree("config.log")

    file, has := conf.String(name)
    if !has {
        file = "log" + string(os.PathSeparator) + name + ".log"
    }

    out, err := createLogfile(file)
    if err == nil {
        logger := log.New(out, "gmvc", 1)
        logger.SetFlags(log.LstdFlags)
        return logger
    }
    return nil
}

func createLogfile(filename string) (*os.File, error) {
    var f *os.File
    var e error
    f, e = os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, os.FileMode(0644))
    if e != nil {
        panic(e)
        return nil, e
    }
    return f, nil
}


