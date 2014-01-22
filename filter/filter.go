package filter

import (
	"errors"
    "bytes"
    "container/list"
)

func ReplaceToChan(pat []byte, rep []byte, buf []byte, ch chan<- []byte) {
    // TODO: reduce variable count: use sliceStart == -1 instead of sliceOpen
    sliceOpen := false
    sliceStart := 0
    bufPos := 0
    // in the prefix of buf where pat might start
    for bufPos <= len(buf) - len(pat) {
        match := bytes.Equal(pat, buf[bufPos:bufPos + len(pat)])
        if match {
            if sliceOpen { // close and send it
                ch <- buf[sliceStart:bufPos]
                sliceOpen = false
            }
            if len(rep) > 0 { // TODO: optimize earlier by checking rep
                ch <- rep // send replacement
            }
            bufPos += len(pat) // advance past match
        } else {
            if !sliceOpen { // open a new slice
                sliceOpen = true
                sliceStart = bufPos
            }
            bufPos += 1
        }
    }
    // send tail of buf, which is not big enough to contain pat
    if sliceOpen {
        ch <- buf[sliceStart:]
    } else {
        ch <- buf[bufPos:]
    }
}

func Collect(ch <-chan []byte) []byte {
    chunks := list.New()
    totalLen := 0
    for chunk := range ch {
        totalLen += len(chunk)
        chunks.PushBack(chunk)
    }
    buf := make([]byte, totalLen)
    bufPos := 0
    for e := chunks.Front(); e != nil ; e = e.Next() {
        chunk := e.Value.([]byte)
        bufPos += copy(buf[bufPos:], chunk)
    }
    return buf
}

func Filter(pat []byte, rep []byte, buf []byte) ([]byte, error) {
    if len(pat) == 0 {
        return nil, errors.New("Pattern must not be empty")
    }
    if len(pat) > len(buf) {
        return buf, nil // pat can't be in buf because it's too big
    }

    ch := make(chan []byte)
    go func() {
        ReplaceToChan(pat, rep, buf, ch)
        close(ch)
    }()
    return Collect(ch), nil
}
