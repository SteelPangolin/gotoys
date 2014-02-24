package arclight

import (
    "time"
    "io"
)

type Node interface {
    Name() string
}

type Dir interface {
    Node
    Children()                ([]Node, error)
    Resolve(relpath string)   (Node,   error)
}

type File interface {
    Node
    Size()      int64
    ModTime()   time.Time
    Reader()    io.Reader
    ReaderAt()  io.ReaderAt
}
