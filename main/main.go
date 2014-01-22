package main

import (
    "fmt"
    "github.com/SteelPangolin/gotoys/filter"
)

func main() {
    pat := []byte("fillory")
    rep := []byte("further")
    buf := []byte("fillorygoat")
    newBuf, err := filter.Filter(pat, rep, buf)
    fmt.Printf("err = %#v\n", err);
    fmt.Printf("newBuf = %#v\n", string(newBuf));
}
