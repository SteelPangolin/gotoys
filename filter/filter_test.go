package filter

import (
    "testing"
    "bytes"
)

func TestEmptyPattern(t *testing.T) {
    pat := []byte("")
    rep := []byte("anything")
    var buf []byte = nil
    _, err := Filter(pat, rep, buf)
    if err == nil {
        t.Errorf("Expected error from Filter(%#v, %#v, %#v)", pat, rep, buf)
    }
}

func TestBufEqualsPat(t *testing.T) {
    pat := []byte("fillory")
    rep := []byte("further")
    buf := []byte("fillory")
    newBuf, err := Filter(pat, rep, buf)
    if err != nil {
        t.Errorf("Unexpected error %#v from Filter(%#v, %#v, %#v)", err, pat, rep, buf)
    }
    if !bytes.Equal(newBuf, rep) {
        t.Errorf("Expected %#v == %#v", string(newBuf), string(rep))
    }
}