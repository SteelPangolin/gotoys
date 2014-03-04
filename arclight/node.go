package arclight

import (
	"io"
	"time"
)

type NodeAttrs map[string]string

type VfsNode interface {
	Name() string
	Attrs() NodeAttrs
}

type VfsDir interface {
	VfsNode
	Children() ([]VfsNode, error)
	Resolve(relpath string) (VfsNode, error)
}

type VfsFile interface {
	VfsNode
	Size() int64
	ModTime() time.Time
	Reader() (io.Reader, error)
}
