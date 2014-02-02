package main

import (
    "log"
    "net/http"
    "encoding/json"
//    "github.com/SteelPangolin/gotoys/filter"
)

func echo(w http.ResponseWriter, r *http.Request) {
    var err error
    var reqJSON []byte
    reqJSON, err = json.MarshalIndent(r, "", "    ")
    if err != nil {
        log.Print(err)
    }
    w.Header().Set("Content-Type", "application/json")
    _, err = w.Write(reqJSON)
    if err != nil {
        log.Print(err)
    }
}

func main() {
    http.HandleFunc("/", echo)
    http.ListenAndServe(":5000", nil)
}
