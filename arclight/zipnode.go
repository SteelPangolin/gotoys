package arclight

import (
    "archive/zip"
)

// The Zip archive is a File and also implements Directory
type ZipArchive struct {
    File
}

// A file inside the archive
type ZipFile struct {
    FileHeader zip.FileHeader
}

// A directory inside the