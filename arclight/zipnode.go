package arclight

import (
	"archive/zip"
	"fmt"
	"io"
	slashpath "path"
	"strings"
	"time"
)

type ZipArchive struct {
	VfsFileNode
}

func NewZipArchive(file VfsFileNode) VfsDirFileNode {
	arc := new(ZipArchive)
	arc.VfsFileNode = file

	// attempt to add Zip file comment to attrs
	reader, err := arc.Reader()
	if err != nil {
		closer, ok := reader.(io.Closer)
		if ok {
			defer closer.Close()
		}

		readerat, ok := reader.(io.ReaderAt)
		if ok {
			z, err := zip.NewReader(readerat, arc.Size())
			if err == nil && z.Comment != "" {
				arc.Attrs()["zip.comment"] = z.Comment
			}
		}
	}

	return arc
}

func (arc *ZipArchive) nodes() ([]zipNode, error) {
	reader, err := arc.Reader()
	if err != nil {
		return nil, err
	}

	closer, ok := reader.(io.Closer)
	if ok {
		defer closer.Close()
	}

	readerat, ok := reader.(io.ReaderAt)
	if !ok {
		err := fmt.Errorf("$T's Reader() doesn't implement readat()",
			arc.VfsFileNode)
		return nil, err
	}

	z, err := zip.NewReader(readerat, arc.Size())
	if err != nil {
		return nil, err
	}

	nodes := make([]zipNode, len(z.File))
	paths := make([]string, len(z.File))
	for i, f := range z.File {
		if f.FileInfo().IsDir() {
			nodes[i] = NewZipDir(arc, &f.FileHeader)
		} else {
			nodes[i] = NewZipFile(f)
		}
		paths[i] = nodes[i].arcPath()
	}

	for _, path := range ImplicitDirs(paths) {
		nodes = append(nodes, NewImplicitZipDir(arc, path))
	}

	return nodes, nil
}

func (arc *ZipArchive) Children() ([]VfsNode, error) {
	nodes, err := arc.nodes()
	if err != nil {
		return nil, err
	}

	children := make([]VfsNode, 0)
	for _, node := range nodes {
		if depth(node) == 1 {
			children = append(children, node)
		}
	}
	return children, nil
}

func (arc *ZipArchive) childrenOf(node zipNode) ([]VfsNode, error) {
	nodes, err := arc.nodes()
	if err != nil {
		return nil, err
	}

	targetDepth := 1 + depth(node)
	pathPfx := node.arcPath() + "/"
	children := make([]VfsNode, 0)

	for _, node := range nodes {
		if depth(node) == targetDepth && strings.HasPrefix(node.arcPath(), pathPfx) {
			children = append(children, node)
		}
	}
	return children, nil
}

func (arc *ZipArchive) Resolve(relpath string) (VfsNode, error) {
	nodes, err := arc.nodes()
	if err != nil {
		return nil, err
	}

	for _, node := range nodes {
		if node.arcPath() == relpath {
			return node, nil
		}
	}

	return nil, fmt.Errorf("Archive path not found", relpath)
}

// All nodes inside the Zip archive have an internal path
type zipNode interface {
	VfsNode
	arcPath() string
}

type zipDir interface {
	arc() *ZipArchive
}

func depth(zn zipNode) int {
	return 1 + strings.Count(slashpath.Clean(zn.arcPath()), "/")
}

// A file inside the archive
type ZipFile struct {
	attrs NodeAttrs
	f     *zip.File
}

func NewZipFile(f *zip.File) *ZipFile {
	node := new(ZipFile)
	node.attrs = make(NodeAttrs)
	if f.Comment != "" {
		node.attrs["zip.comment"] = f.Comment
	}
	node.f = f
	return node
}

func (node *ZipFile) arcPath() string {
	return slashpath.Clean(node.f.Name)
}

func (node *ZipFile) Attrs() NodeAttrs {
	return node.attrs
}

func (node *ZipFile) Name() string {
	return node.f.FileInfo().Name()
}

func (node *ZipFile) Size() int64 {
	return node.f.FileInfo().Size()
}

func (node *ZipFile) ModTime() time.Time {
	return node.f.FileInfo().ModTime()
}

func (node *ZipFile) MimeType() string {
	return OctetStream // TODO
}

func (node *ZipFile) Reader() (io.Reader, error) {
	return node.f.Open()
}

// A directory inside the archive
type ZipDir struct {
	attrs NodeAttrs
	arc   *ZipArchive
	fh    *zip.FileHeader
}

func NewZipDir(arc *ZipArchive, fh *zip.FileHeader) *ZipDir {
	node := new(ZipDir)
	node.attrs = make(NodeAttrs)
	if fh.Comment != "" {
		node.attrs["zip.comment"] = fh.Comment
	}
	node.arc = arc
	node.fh = fh
	return node
}

func (node *ZipDir) arcPath() string {
	return slashpath.Clean(node.fh.Name)
}

func (node *ZipDir) Attrs() NodeAttrs {
	return node.attrs
}

func (node *ZipDir) Name() string {
	return node.fh.FileInfo().Name()
}

func (node *ZipDir) ModTime() time.Time {
	return node.fh.FileInfo().ModTime()
}

func (node *ZipDir) MimeType() string {
	return InodeDirectory
}

func (node *ZipDir) Children() ([]VfsNode, error) {
	return node.arc.childrenOf(node)
}

func (node *ZipDir) Resolve(relpath string) (VfsNode, error) {
	return node.arc.Resolve(slashpath.Join(node.fh.Name, relpath))
}

// A directory not present in the Zip archive,
// but implied by other entries with paths of
// which this directory's path is a prefix.
type ImplicitZipDir struct {
	attrs NodeAttrs
	arc   *ZipArchive
	path  string
}

func NewImplicitZipDir(arc *ZipArchive, path string) *ImplicitZipDir {
	node := new(ImplicitZipDir)
	node.attrs = make(NodeAttrs)
	node.arc = arc
	node.path = path
	return node
}

func (node *ImplicitZipDir) arcPath() string {
	// already cleaned
	return node.path
}

func (node *ImplicitZipDir) Attrs() NodeAttrs {
	return node.attrs
}

func (node *ImplicitZipDir) Name() string {
	return slashpath.Base(node.path)
}

func (node *ImplicitZipDir) ModTime() time.Time {
	return node.arc.ModTime()
}

func (node *ImplicitZipDir) MimeType() string {
	return InodeDirectory
}

func (node *ImplicitZipDir) Children() ([]VfsNode, error) {
	return node.arc.childrenOf(node)
}

func (node *ImplicitZipDir) Resolve(relpath string) (VfsNode, error) {
	return node.arc.Resolve(slashpath.Join(node.path, relpath))
}
