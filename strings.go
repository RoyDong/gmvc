package gmvc

import (
    "bytes"
    "encoding/json"
    "crypto/rand"
    "os"
    "regexp"
    yaml "gopkg.in/yaml.v2"
    "crypto/md5"
    "io"
    "encoding/hex"
)

/*
LoadJson reads data from a json format file to v
*/
func LoadJson(v interface{}, filename string) error {
    text, err := LoadFile(filename)
    if err != nil {
        return err
    }

    rows := bytes.Split(text, []byte("\n"))
    r := regexp.MustCompile(`^\s*[/#]+`)
    for i, row := range rows {
        if r.Match(row) {
            rows[i] = nil
        }
    }

    return json.Unmarshal(bytes.Join(rows, nil), v)
}


/**
LoadYaml reads data from a yaml format file to v
*/
func LoadYaml(v interface{}, filename string) error {
    text, err := LoadFile(filename)
    if err != nil {
        return err
    }

    return yaml.Unmarshal(text, v)
}


/*
LoadJson reads data from a file and returns it as bytes
*/
func LoadFile(filename string) ([]byte, error) {
    file, err := os.Open(filename)
    if err != nil {
        return nil, err
    }
    defer file.Close()

    fileInfo, err := file.Stat()
    if err != nil {
        return nil, err
    }

    text := make([]byte, fileInfo.Size())
    file.Read(text)
    return text, nil
}

func MD5(txt string) string {
    hash := md5.New()
    if _, err := io.WriteString(hash, txt); err != nil {
        Logger.Fatalln(err.Error())
    }
    return hex.EncodeToString(hash.Sum(nil))
}


/*
RandString creates a random string has length letters
letters are all from Chars
*/
func RandString(n int) string {
    rnd := make([]byte, n)
    if _, err := io.ReadFull(rand.Reader, rnd); err != nil {
        Logger.Fatalln(err.Error())
    }

    return string(rnd)
}

