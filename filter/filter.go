package filter

import (
	"errors"
    "bytes"
    "container/list"
)

const noSlice = -1 // slice is not open

func ReplaceToChan(pat []byte, rep []byte, buf []byte, out chan<- []byte, tailReturn chan<- []byte) {
    sliceStart := noSlice
    bufPos := 0
    // positions where pat can start and fit entirely within buf 
    for bufPos <= len(buf) - len(pat) {
        match := bytes.Equal(pat, buf[bufPos:bufPos + len(pat)])
        if match {
            if sliceStart != noSlice { // slice is open
                // close and send it
                out <- buf[sliceStart:bufPos]
                sliceStart = noSlice
            }
            if len(rep) > 0 { // TODO: optimize earlier by checking rep
                out <- rep // send replacement
            }
            bufPos += len(pat) // advance past match
        } else {
            if sliceStart == noSlice { // slice is not open
                // open it
                sliceStart = bufPos
            }
            bufPos += 1
        }
    }
    // does the unscanned tail of buf end with a prefix of pat?
    // TODO: optimize for bufPos = len(buf)
    tail := buf[bufPos:]
    prefixMatchAtEnd := bytes.Equal(tail, pat[:-len(tail)])
    if sliceStart != noSlice { // slice is open
        if prefixMatchAtEnd {
            // last slice should not include tail
            out <- buf[sliceStart:bufPos]
        } else {
            // last slice goes all the way to the end of buf
            out <- buf[sliceStart:]
        }
    }
    tailReturn <- tail
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

func Finish(out chan<- []byte, tailReturn <-chan []byte) {
    tail := <- tailReturn
    out <- tail
    close(out)
}

func Filter(pat []byte, rep []byte, buf []byte) ([]byte, error) {
    if len(pat) == 0 {
        return nil, errors.New("Pattern must not be empty")
    }
    if len(pat) > len(buf) {
        return buf, nil // pat can't be in buf because it's too big
    }

    out := make(chan []byte)
    tailReturn := make(chan []byte)
    go ReplaceToChan(pat, rep, buf, out, tailReturn)
    go Finish(out, tailReturn)
    return Collect(out), nil
}
