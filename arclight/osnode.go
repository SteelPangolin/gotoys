package arclight

import (
	"io"
	"os"
	slashpath "path"
)

type OsNode struct {
	attrs NodeAttrs
	Path  string
	os.FileInfo
}

func (node *OsNode) Attrs() NodeAttrs {
	return node.attrs
}

func NewOsNode(path string, fi os.FileInfo) VfsNode {
	if fi.IsDir() {
		return NewOsDir(path, fi)
	} else {
		return NewOsFile(path, fi)
	}
}

type OsDir struct {
	OsNode
}

func NewOsDir(path string, fi os.FileInfo) *OsDir {
	node := new(OsDir)
	node.attrs = make(NodeAttrs)
	node.Path = path
	node.FileInfo = fi
	return node
}

func (dir *OsDir) Children() ([]VfsNode, error) {
	f, err := os.Open(dir.Path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	fis, err := f.Readdir(0)
	if err != nil {
		return nil, err
	}

	children := make([]VfsNode, len(fis))
	for i, fi := range fis {
		path := slashpath.Join(dir.Path, fi.Name())
		if fi.IsDir() {
			children[i] = NewOsDir(path, fi)
		} else {
			children[i] = NewOsFile(path, fi)
		}
	}

	return children, nil
}

func (dir *OsDir) Resolve(relpath string) (VfsNode, error) {
	path := slashpath.Join(dir.Path, relpath)
	fi, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	var node VfsNode
	if fi.IsDir() {
		node = NewOsDir(path, fi)
	} else {
		node = NewOsFile(path, fi)
	}

	return node, nil
}

func (dir *OsDir) MimeType() string {
	return InodeDirectory
}

type OsFile struct {
	OsNode
}

func NewOsFile(path string, fi os.FileInfo) *OsFile {
	node := new(OsFile)
	node.attrs = make(NodeAttrs)
	node.Path = path
	node.FileInfo = fi
	return node
}

func (file *OsFile) Reader() (io.Reader, error) {
	return os.Open(file.Path)
}

func (file *OsFile) MimeType() string {
	// special file types
	if mimetype := file.inodeMimeType(); mimetype != OctetStream {
		return mimetype
	}

	mimetype := MagicMimeTypeFromFile(file.Path)
	return mimetype
}

type modeMime struct {
	mode os.FileMode
	mime string
}

// Order is significant: in Go, all CharDevices are also Devices.
var modeMimes = []modeMime{
	{
		mode: os.ModeCharDevice,
		mime: "inode/chardevice",
	},
	{
		mode: os.ModeDevice,
		mime: "inode/blockdevice",
	},
	{
		mode: os.ModeNamedPipe,
		mime: "inode/fifo",
	},
	{
		mode: os.ModeSocket,
		mime: "inode/socket",
	},
}

func (file *OsFile) inodeMimeType() string {
	mode := file.Mode()
	for _, entry := range modeMimes {
		if mode&entry.mode == entry.mode {
			return entry.mime
		}
	}
	return OctetStream
}
