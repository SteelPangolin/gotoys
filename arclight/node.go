package arclight

import (
	"io"
	"time"
)

type NodeAttrs map[string]string

type VfsNode interface {
	Name() string
	ModTime() time.Time
	Attrs() NodeAttrs
	MimeType() (string, map[string]string)
}

type VfsDir interface {
	Children() ([]VfsNode, error)
	Resolve(relpath string) (VfsNode, error)
}

type VfsDirNode interface {
	VfsNode
	VfsDir
}

type VfsFile interface {
	Size() int64
	Open() (io.ReadCloser, error)
}

type VfsFileNode interface {
	VfsNode
	VfsFile
}

type VfsDirFileNode interface {
	VfsNode
	VfsDir
	VfsFile
}
