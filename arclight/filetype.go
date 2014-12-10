package arclight

import (
	"errors"
	"github.com/rakyll/magicmime"
	"io"
	"log"
	"mime"
	"os"
	slashpath "path"
	"strings"
)

type modeMime struct {
	mode os.FileMode
	mime string
}

// Order is significant: in Go, all CharDevices are also Devices.
var modeMimes = []modeMime{
	{
		mode: os.ModeDir,
		mime: "inode/directory",
	},
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

var OctetStream = "application/octet-stream"

var RealFSOnly = errors.New("Must be on a real filesystem")

// ignores symlinks
// see http://standards.freedesktop.org/shared-mime-info-spec/shared-mime-info-spec-latest.html#idm140625828597376
func MimeTypeByStat(node VfsNode) (string, error) {
	osnode, ok := node.(*OsNode) // TODO: is this a legit cast?
	if !ok {
		return "", RealFSOnly
	}
	mode := osnode.Mode()
	for _, entry := range modeMimes {
		if mode&entry.mode == entry.mode {
			return entry.mime, nil
		}
	}
	return OctetStream, nil
}

// Detect MIME type using file name extension.
func MimeTypeByExt(node VfsNode) (string, error) {
	mimetype_raw := mime.TypeByExtension(slashpath.Ext(node.Name()))

	mediatype, params, err := mime.ParseMediaType(mimetype_raw)
	if err != nil {
		return "", err
	}

	// mime package lies about text type charsets, so drop them
	if strings.HasPrefix(mediatype, "text/") {
		delete(params, "charset")
	}

	mimetype := mime.FormatMediaType(mediatype, params)

	return mimetype, nil
}

// Initialized the first time MimeTypeByMagic() is called.
var magic *magicmime.Magic
var magic_err error

func MimeTypeByMagic_recover() {
	if r := recover(); r != nil {
		log.Printf("%#q\n", r)
	}
}

// Detect MIME type using libmagic
func MimeTypeByMagic(node VfsNode) (string, error) {
	defer MimeTypeByMagic_recover()

	// init libmagic on first call
	// TODO: lock? what happens if it gets inited twice?
	if magic == nil {
		magic, magic_err = magicmime.New(magicmime.MAGIC_MIME)
	}
	if magic_err != nil {
		return "", magic_err
	}

	var mimetype_raw string
	var err error
	osfile, ok := node.(*OsFile)
	if ok {
		mimetype_raw, err = magic.TypeByFile(osfile.Path)
	} else {
		file, ok := node.(VfsFile)
		if !ok {
			return "", errors.New("not a file, can't look for magic")
		}
		// read up to 512 bytes
		// enough for most MIME sniffers?
		reader, err := file.Reader()
		if err != nil {
			return "", err
		}

		closer, ok := reader.(io.Closer)
		if ok {
			defer closer.Close()
		}

		buf := make([]byte, 512)
		n, err := reader.Read(buf)
		if err != nil && err != io.EOF {
			return "", err
		}
		buf = buf[:n]

		mimetype_raw, err = magic.TypeByBuffer(buf)
	}
	if err != nil {
		return "", err
	}

	mediatype, params, err := mime.ParseMediaType(mimetype_raw)
	if err != nil {
		return "", err
	}

	// We don't care if it's empty, it should still use a standard MIME type.
	if mediatype == "inode/x-empty" {
		mediatype = OctetStream
	}

	// There is no binary charset, even for types that actually have a charset param.
	// It's an annoying artifact of libmagic.
	if params["charset"] == "binary" {
		delete(params, "charset")
	}

	mimetype := mime.FormatMediaType(mediatype, params)

	return mimetype, nil
}

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
