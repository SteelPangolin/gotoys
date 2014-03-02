package arclight

import (
	"github.com/rakyll/magicmime"
	"io"
	"log"
	"mime"
	slashpath "path"
)

// Detect MIME type using file name extension.
func MimeTypeByExt(file VfsFile) string {
	return mime.TypeByExtension(slashpath.Ext(file.Name()))
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
func MimeTypeByMagic(file VfsFile) (string, error) {
	defer MimeTypeByMagic_recover()

	// init libmagic on first call
	if magic == nil {
		magic, magic_err = magicmime.New()
	}
	if magic_err != nil {
		return "", magic_err
	}

	var mimetype_raw string
	var err error
	osfile, ok := file.(*OsFile)
	if ok {
		mimetype_raw, err = magic.TypeByFile(osfile.Path)
	} else {
		// read up to 512 bytes
		// enough for most MIME sniffers?
		reader, err := file.Reader()
		if err != nil {
			return "", err
		}

		buf := make([]byte, 512)
		n, err := reader.Read(buf)
		if err != nil && err != io.EOF {
			return "", err
		}
		buf = buf[:n]

		readcloser, ok := reader.(io.ReadCloser)
		if ok {
			readcloser.Close()
		}

		mimetype_raw, err = magic.TypeByBuffer(buf)
	}
	if err != nil {
		return "", err
	}

	mediatype, params, err := mime.ParseMediaType(mimetype_raw)
	if err != nil {
		return "", err
	}

	mimetype := mime.FormatMediaType(mediatype, params)

	return mimetype, nil
}

func DetectMimeTypes(file VfsFile) {
	mimetype_ext := MimeTypeByExt(file)

	mimetype_magic, err := MimeTypeByMagic(file)
	if err != nil {
		mimetype_magic = ""
	}

	// TODO: resolve conflicts when mediatype is the same but
	// params are only present on one

	if mimetype_ext == "" && mimetype_magic == "" {
		return
	}

	if mimetype_ext == mimetype_magic {
		file.Attrs()["mimetype"] = mimetype_ext
		return
	}

	if mimetype_ext != "" && mimetype_magic != "" {
		file.Attrs()["mimetype_ext"] = mimetype_ext
		file.Attrs()["mimetype_magic"] = mimetype_magic
		return
	}

	if mimetype_ext != "" {
		file.Attrs()["mimetype"] = mimetype_ext
		return
	}

	if mimetype_magic != "" {
		file.Attrs()["mimetype"] = mimetype_magic
		return
	}
}
