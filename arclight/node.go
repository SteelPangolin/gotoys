package arclight

import (
	"io"
	"time"
)

type VfsNode interface {
	Name() string
	Attrs() map[string]string
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
