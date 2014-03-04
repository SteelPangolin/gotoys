package arclight

import (
	"archive/zip"
	"fmt"
	"io"
    "time"
    slashpath "path"
)

type ZipArchive struct {
	VfsFile
}

func (arc *ZipArchive) nodes() ([]VfsNode, error) {
	reader, err := arc.Reader()
	if err != nil {
		return nil, err
	}

	readerat, ok := reader.(io.ReaderAt)
	if !ok {
		err := fmt.Errorf("$T's Reader() doesn't implement readat()",
			arc.VfsFile)
		return nil, err
	}

	z, err := zip.NewReader(readerat, arc.Size())
	if err != nil {
		return nil, err
	}

	nodes := make([]VfsNode, len(z.File))
	paths := make([]string, len(z.File))
	for i, f := range z.File {
		paths[i] = f.Name
        if f.FileInfo().IsDir() {

        } else {

        }
	}

    for _, path := range ImplicitDirs(paths) {
        nodes = append(nodes, NewImplicitZipDir(arc, path))
    }

	return nodes, nil
}

func (arc *ZipArchive) Children() ([]VfsNode, error) {
	// TODO: depth check: show only depth = 1 children
	return arc.nodes()
}

func (arc *ZipArchive) Resolve(relpath string) (VfsNode, error) {
    // TODO: implement
	nodes, err := arc.nodes()
    if err != nil || len(nodes) == 0 {
        return nil, err
    } else {
        return nodes[0], nil
    }
}

// A file inside the archive
type ZipFile struct {
    attrs      NodeAttrs
    file       *zip.File
}

func NewZipFile(f *zip.File) VfsFile {
    node := new(ZipFile)
    node.attrs = make(NodeAttrs)
    if f.Comment != "" {
        node.attrs["zip.comment"] = f.Comment
    }
    node.file = f
    return node
}

func (node *ZipFile) Attrs() NodeAttrs {
    return node.attrs
}

func (node *ZipFile) Name() string {
    return node.file.FileInfo().Name()
}

func (node *ZipFile) Size() int64 {
    return node.file.FileInfo().Size()
}

func (node *ZipFile) ModTime() time.Time {
    return node.file.FileInfo().ModTime()
}

func (node *ZipFile) Reader() (io.Reader, error) {
    return node.file.Open()
}

// A directory inside the archive
type ZipDir struct {
    attrs      NodeAttrs
    arc        *ZipArchive
    path       string
}

func NewZipDir(arc *ZipArchive, fh *zip.FileHeader) VfsDir {
    node := new(ZipDir)
    node.attrs = make(NodeAttrs)
    if fh.Comment != "" {
        node.attrs["zip.comment"] = fh.Comment
    }
    node.arc = arc
    node.path = fh.Name
    return node
}

func (node *ZipDir) Attrs() NodeAttrs {
    return node.attrs
}

func (node *ZipDir) Name() string {
    return slashpath.Base(node.path)
}

func (node *ZipDir) Children() ([]VfsNode, error) {
    // TODO
    return node.arc.Children()
}

func (node *ZipDir) Resolve(relpath string) (VfsNode, error) {
    // TODO
    return node.arc.Resolve(relpath)
}

// A directory not present in the Zip archive,
// but implied by other entries with paths of
// which this directory's path is a prefix.
type ImplicitZipDir struct {
    attrs      NodeAttrs
    arc        *ZipArchive
    path       string
}

func NewImplicitZipDir(arc *ZipArchive, path string) VfsDir {
    node := new(ImplicitZipDir)
    node.attrs = make(NodeAttrs)
    node.arc = arc
    node.path = path
    return node
}

func (node *ImplicitZipDir) Attrs() NodeAttrs {
    return node.attrs
}

func (node *ImplicitZipDir) Name() string {
    return slashpath.Base(node.path)
}

func (node *ImplicitZipDir) Children() ([]VfsNode, error) {
    // TODO
    return node.arc.Children()
}

func (node *ImplicitZipDir) Resolve(relpath string) (VfsNode, error) {
    // TODO
    return node.arc.Resolve(relpath)
}