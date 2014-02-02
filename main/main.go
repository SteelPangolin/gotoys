package main

import (
    "log"
    "bytes"
    "github.com/SteelPangolin/gotoys/filter"
)

func main() {
    pat := []byte("fillory")
    rep := []byte("further")
    buf := []byte("fillorygoats")
    expected := []byte("furthergoats")
    newBuf, err := filter.Filter(pat, rep, buf)
    if err != nil {
        log.Printf("Unexpected error %#v from Filter(%#v, %#v, %#v)", err, pat, rep, buf)
    }
    if !bytes.Equal(newBuf, rep) {
        log.Printf("Expected %#v == %#v", string(newBuf), string(rep))
    }
}
