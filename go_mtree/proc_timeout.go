package main

import (
    "os/exec"
    "time"
    "fmt"
)

func externalProc() chan string {
    ch := make(chan string)
    go func() {
        cmd := exec.Command("bash", "-c", "sleep 1 && echo Cougar Boost!")
        out, _ := cmd.Output()
        ch <- string(out)
    }()
    return ch
}

func main() {
    select {
    case out := <- externalProc():
        fmt.Println(out)
    case <- time.After(2 * time.Second):
        fmt.Println("timeout")
    }
}