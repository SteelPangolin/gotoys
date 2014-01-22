package filter

import (
    "testing"
    "bytes"
)

func TestEmptyPattern(t *testing.T) {
    pat := []byte("")
    rep := []byte("anything")
    var buf []byte = nil
    err := ReplaceInPlace(pat, rep, buf)
    if err == nil {
        t.Errorf("Expected error from ReplaceInPlace(%#v, %#v, %#v)", pat, rep, buf)
    }
}

func TestDifferentLengths(t *testing.T) {
    pat := []byte("halibut")
    rep := []byte("anything")
    var buf []byte = nil
    err := ReplaceInPlace(pat, rep, buf)
    if err == nil {
        t.Errorf("Expected error from ReplaceInPlace(%#v, %#v, %#v)", pat, rep, buf)
    }
}

func TestBufEqualsPat(t *testing.T) {
    pat := []byte("fillory")
    rep := []byte("further")
    buf := []byte("fillory")
    err := ReplaceInPlace(pat, rep, buf)
    if err != nil {
        t.Errorf("Unexpected error %#v from ReplaceInPlace(%#v, %#v, %#v)", err, pat, rep, buf)
    }
    if !bytes.Equal(buf, rep) {
        t.Errorf("Expected %#v == %#v", string(buf), string(rep))
    }
}