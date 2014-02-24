package arclight

import (
    "os"
    slashpath "path"
)

type OsNode struct {
    Path string
    Info os.FileInfo
}

func (node *OsNode) Name() string {
    return node.Info.Name()
}

type OsDir struct {
    OsNode
}

func NewOsNode(path string, fi os.FileInfo) Node {
    var node Node
    if fi.IsDir() {
        dir := new(OsDir)
        dir.Path = path
        dir.Info = fi
        node = dir
    } else {
        file := new(OsFile)
        file.Path = path
        file.Info = fi
        node = file
    }
    return Specialize(node)
}

func (dir *OsDir) Children() ([]Node, error) {
    f, err := os.Open(dir.Path)
    if err != nil {
        return nil, err
    }
    fis, err := f.Readdir(0)
    if err != nil {
        return nil, err
    }
    children := make([]Node, len(fis))
    for i, fi := range fis {
        path := slashpath.Join(dir.Path, fi.Name())
        children[i] = NewOsNode(path, fi)
    }
    return children, nil
}

func (dir *OsDir) Resolve(relpath string) (Node, error) {
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

func (file *OsFile) Size() int64 {
    return file.Info.Size()
}

func (file *OsFile) ModTime() time.Time {
    return file.Info.Size()
}
