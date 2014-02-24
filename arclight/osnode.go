package arclight

import (
    "io"
    "os"
    slashpath "path"
)


type OsNode struct {
    attrs map[string]string
    Path string
    os.FileInfo
}

func (node *OsNode) Name() string {
    return node.FileInfo.Name()
}

func (node *OsNode) Attrs() map[string]string {
    return node.attrs
}

func NewOsNode(path string, fi os.FileInfo) VfsNode {
    var node VfsNode
    if fi.IsDir() {
        dir := new(OsDir)
        dir.attrs = make(map[string]string)
        dir.Path = path
        dir.FileInfo = fi
        node = dir
    } else {
        file := new(OsFile)
        file.attrs = make(map[string]string)
        file.Path = path
        file.FileInfo = fi
        node = file
    }
    return Specialize(node)
}


type OsDir struct {
    OsNode
}

func (dir *OsDir) Children() ([]VfsNode, error) {
    f, err := os.Open(dir.Path)
    if err != nil {
        return nil, err
    }

    fis, err := f.Readdir(0)
    if err != nil {
        return nil, err
    }

    children := make([]VfsNode, len(fis))
    for i, fi := range fis {
        path := slashpath.Join(dir.Path, fi.Name())
        children[i] = NewOsNode(path, fi)
    }

    return children, nil
}

func (dir *OsDir) Resolve(relpath string) (VfsNode, error) {
    path := slashpath.Join(dir.Path, relpath)
    fi, err := os.Stat(path)
    if err != nil {
        return nil, err
    }

    node := NewOsNode(path, fi)

    return node, nil
}


type OsFile struct {
    OsNode
}

func (file *OsFile) Reader() (io.Reader, error) {
    return os.Open(file.Path)
}
