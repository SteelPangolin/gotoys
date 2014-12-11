package arclight

import (
	"github.com/rakyll/magicmime"
	"io"
	"log"
	"mime"
	slashpath "path"
	"strings"
)

const (
	InodeDirectory = "inode/directory"
	OctetStream    = "application/octet-stream"
)

// Detect MIME type using file name extension.
func MimeTypeByExt(path string) string {
	mimetype := mime.TypeByExtension(slashpath.Ext(path))

	mediatype, params, err := mime.ParseMediaType(mimetype)
	if err != nil {
		// mime.TypeByExtension() should always return a properly formed MIME type,
		// so this should never happen.
		log.Printf("WARNING: mime.TypeByExtension() returned MIME type %#v", mimetype)
		return OctetStream
	}

	// mime package lies about text type charsets, so drop them
	if strings.HasPrefix(mediatype, "text/") {
		delete(params, "charset")
	}

	return mime.FormatMediaType(mediatype, params)
}

var MimeTypeFromFile func(path string) string = StubMimeTypeFromFile
var MimeTypeFromReader func(open func() (io.Reader, error)) string = StubMimeTypeFromReader

func StubMimeTypeFromFile(path string) string {
	return OctetStream
}

func StubMimeTypeFromReader(open func() (io.Reader, error)) string {
	return OctetStream
}

// Interface to libmagic.
var magic *magicmime.Magic

// Initialize libmagic.
func init() {
	var err error
	magic, err = magicmime.New(magicmime.MAGIC_MIME)
	if err != nil {
		return
	}
	MimeTypeFromFile = MagicMimeTypeFromFile
	MimeTypeFromReader = MagicMimeTypeFromReader
}

func MagicMimeTypeFromFile(path string) string {
	mimetype, err := magic.TypeByFile(path)
	if err != nil {
		log.Printf("WARNING: libmagic error: %v", err)
		return OctetStream
	}
	return cleanupMimeTypeByMagic(mimetype)
}

func MagicMimeTypeFromReader(open func() (io.Reader, error)) string {
	reader, err := open()
	if err != nil {
		log.Printf("WARNING: couldn't open reader: %v", err)
		return OctetStream
	}

	if closer, ok := reader.(io.Closer); ok {
		defer closer.Close()
	}

	// read up to 512 bytes
	// enough for most file types?
	numBytes := 512
	buf := make([]byte, numBytes)
	n, err := reader.Read(buf)
	if err != nil && err != io.EOF {
		log.Printf("WARNING: error while trying to read %d bytes: %v", numBytes, err)
		return OctetStream
	}
	buf = buf[:n]

	mimetype, err := magic.TypeByBuffer(buf)
	if err != nil {
		log.Printf("WARNING: libmagic error: %v", err)
		return OctetStream
	}
	return cleanupMimeTypeByMagic(mimetype)
}

// libmagic isn't always helpful.
func cleanupMimeTypeByMagic(mimetype string) string {
	mediatype, params, err := mime.ParseMediaType(mimetype)
	if err != nil {
		// libmagic should always return a properly formed MIME type,
		// so this should never happen.
		log.Printf("WARNING: libmagic returned improperly formed MIME type %#v", mimetype)
		return OctetStream
	}

	// We don't care if it's empty, it should still use a standard MIME type.
	if mediatype == "inode/x-empty" {
		return OctetStream
	}

	// There is no binary charset, even for types that actually have a charset param.
	// It's an annoying artifact of libmagic.
	if params["charset"] == "binary" {
		delete(params, "charset")
	}

	return mime.FormatMediaType(mediatype, params)
}

// TODO: generalize for any node that has both a file name and contents.
/*
func DetectMimeTypes(node VfsNode) {
	// this will only assign MIME types to stuff that isn't a regular file
	mimetype_stat, err := MimeTypeByStat(node)
	if err == nil && mimetype_stat != OctetStream {
		node.Attrs()["mimetype"] = mimetype_stat
		return
	}

	mimetype_ext, err := MimeTypeByExt(node)
	if err != nil {
		mimetype_ext = OctetStream
	}

	mimetype_magic, err := MimeTypeByMagic(node)
	if err != nil {
		mimetype_magic = OctetStream
	}

	// TODO: resolve conflicts when mediatype is the same but
	// params are only present on one

	if mimetype_ext == OctetStream && mimetype_magic == OctetStream {
		return
	}

	if mimetype_ext == mimetype_magic {
		node.Attrs()["mimetype"] = mimetype_ext
		return
	}

	if mimetype_ext != OctetStream && mimetype_magic != OctetStream {
		node.Attrs()["mimetype_ext"] = mimetype_ext
		node.Attrs()["mimetype_magic"] = mimetype_magic
		return
	}

	if mimetype_ext != OctetStream {
		node.Attrs()["mimetype"] = mimetype_ext
		return
	}

	if mimetype_magic != OctetStream {
		node.Attrs()["mimetype"] = mimetype_magic
		return
	}
}
*/
