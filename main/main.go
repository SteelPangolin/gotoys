package main

import (
    "fmt"
    "github.com/SteelPangolin/gotoys/filter"
)

func main() {
    pat := []byte("fillory")
    rep := []byte("further")
    buf := []byte("fillory")
    err := filter.ReplaceInPlace(pat, rep, buf)
    fmt.Printf("err = %#v\n", err);
    fmt.Printf("buf = %#v\n", string(buf));
}