package arclight

import (
	"archive/zip"
	//    slashpath "path"
	"fmt"
	"io"
	"log"
)

type ZipArchive struct {
	VfsFile
}

type nodeFilter func(f zip.File) bool

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
	nodePaths := make([]string, len(z.File))
	for i, f := range z.File {
		nodes[i] = NewZipNode(f.FileHeader)
		nodePaths[i] = f.FileHeader.Name
	}

	return nodes, nil
}

func (arc *ZipArchive) Children() ([]VfsNode, error) {
	// TODO: depth check: show only depth = 1 children
	return arc.nodes()
}

func (arc *ZipArchive) Resolve(relpath string) (VfsNode, error) {
	log.Panic("Resolve() not implemented for ZipArchive")
	return nil, nil

	/*
	   path := slashpath.Join(dir.Path, relpath)
	   fi, err := os.Stat(path)
	   if err != nil {
	       return nil, err
	   }
	   node := NewOsNode(path, fi)
	   return node, nil
	*/
}

type ZipNode struct {
	attrs      map[string]string
	Archive    ZipArchive
	FileHeader zip.FileHeader
}

func (node *ZipNode) Name() string {
	return node.FileHeader.FileInfo().Name()
}

func (node *ZipNode) Attrs() map[string]string {
	return node.attrs
}

func NewZipNode(fh zip.FileHeader) VfsNode {
	var node VfsNode
	fi := fh.FileInfo()
	if fi.IsDir() {
		dir := new(ZipDir)
		dir.attrs = make(map[string]string)
		dir.FileHeader = fh
		node = dir
	} else {
		file := new(ZipFile)
		file.attrs = make(map[string]string)
		file.FileHeader = fh
		node = file
	}
	if fh.Comment != "" {
		node.Attrs()["zip_comment"] = fh.Comment
	}
	return Specialize(node)
}

// A file inside the archive
type ZipFile struct {
	ZipNode
}

// A directory inside the archive
type ZipDir struct {
	ZipNode
}

// A directory not present in the Zip archive,
// but implied by other entries with paths of
// which this directory's path is a prefix.
type ImplicitZipDir struct {
	ZipNode
}
