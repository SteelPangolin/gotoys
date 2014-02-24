package arclight

func Specialize(orig VfsNode) VfsNode {
    if file, ok := orig.(VfsFile); ok {
        DetectMimeTypes(file)
        if mimetype, ok := file.Attrs()["mimetype"]; ok {
            if mimetype == "application/zip" {
                arc := new(ZipArchive)
                arc.VfsFile = file
                return arc
            }
        }
    }
    return orig
}
