package gmvc

import (
    "bytes"
    "encoding/json"
    "math/rand"
    "os"
    "regexp"
)

const (
    Chars = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
)

/*
RandString creates a random string has length letters
letters are all from Chars
*/
func RandString(length int) string {
    rnd := make([]byte, 0, length)
    max := len(Chars) - 1

    for i := 0; i < length; i++ {
        rnd = append(rnd, Chars[rand.Intn(max)])
    }

    return string(rnd)
}

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


/*
LoadYaml reads data from a yaml format file to v
func LoadYaml(v interface{}, filename string) error {
    text, err := LoadFile(filename)
    if err != nil {
        return err
    }

    return goyaml.Unmarshal(text, v)
}
*/


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
