package step

import (
    "net/http"
    "fmt"
    "bytes"
    "io/ioutil"
)

func HTTPSend(url string, data []byte) {
    reader := bytes.NewReader(data)
    request, err := http.NewRequest("POST", url, reader)
    if err != nil {
        fmt.Println(err.Error())
        return
    }
    request.Header.Set("Content-Type", "application/json;charset=UTF-8")
    client := http.Client{}
    resp, err := client.Do(request)
    if err != nil {
        fmt.Println(err.Error())
        return
    }
    respBytes, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        fmt.Println(err.Error())
        return
    }

    fmt.Println(string(respBytes))
}
