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
func MimeTypeByExt(path string) (string, map[string]string) {
	mimetype := mime.TypeByExtension(slashpath.Ext(path))

	mediatype, params, err := mime.ParseMediaType(mimetype)
	if err != nil {
		// mime.TypeByExtension() should always return a properly formed MIME type,
		// so this should never happen.
		log.Printf("WARNING: mime.TypeByExtension() returned MIME type %#v", mimetype)
		return OctetStream, nil
	}

	// mime package lies about text type charsets, so drop them
	if strings.HasPrefix(mediatype, "text/") {
		delete(params, "charset")
	}

	return mediatype, params
}

var MimeTypeFromFile func(path string) (string, map[string]string) = StubMimeTypeFromFile
var MimeTypeFromReader func(open func() (io.ReadCloser, error)) (string, map[string]string) = StubMimeTypeFromReader

func StubMimeTypeFromFile(path string) (string, map[string]string) {
	return OctetStream, nil
}

func StubMimeTypeFromReader(open func() (io.ReadCloser, error)) (string, map[string]string) {
	return OctetStream, nil
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

func MagicMimeTypeFromFile(path string) (string, map[string]string) {
	mimetype, err := magic.TypeByFile(path)
	if err != nil {
		log.Printf("WARNING: libmagic error: %v", err)
		return OctetStream, nil
	}
	mediatype, params := cleanupMimeTypeByMagic(mimetype)
	return mediatype, params
}

func MagicMimeTypeFromReader(open func() (io.ReadCloser, error)) (string, map[string]string) {
	reader, err := open()
	if err != nil {
		log.Printf("WARNING: couldn't open reader: %v", err)
		return OctetStream, nil
	}
	defer reader.Close()

	// read up to 512 bytes
	// enough for most file types?
	numBytes := 512
	buf := make([]byte, numBytes)
	n, err := reader.Read(buf)
	if err != nil && err != io.EOF {
		log.Printf("WARNING: error while trying to read %d bytes: %v", numBytes, err)
		return OctetStream, nil
	}
	buf = buf[:n]

	mimetype, err := magic.TypeByBuffer(buf)
	if err != nil {
		log.Printf("WARNING: libmagic error: %v", err)
		return OctetStream, nil
	}
	mediatype, params := cleanupMimeTypeByMagic(mimetype)
	return mediatype, params
}

// libmagic isn't always helpful.
func cleanupMimeTypeByMagic(mimetype string) (string, map[string]string) {
	mediatype, params, err := mime.ParseMediaType(mimetype)
	if err != nil {
		// libmagic should always return a properly formed MIME type,
		// so this should never happen.
		// One exception is if we don't have permissions to read the file.
		log.Printf("WARNING: libmagic returned improperly formed MIME type %#v", mimetype)
		return OctetStream, nil
	}

	// We don't care if it's empty, it should still use a standard MIME type.
	if mediatype == "inode/x-empty" {
		return OctetStream, nil
	}

	// There is no binary charset, even for types that actually have a charset param.
	// It's an annoying artifact of libmagic.
	if params["charset"] == "binary" {
		delete(params, "charset")
	}

	return mediatype, params
}
