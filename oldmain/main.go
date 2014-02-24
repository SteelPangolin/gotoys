package main

import (
    "os"
    "log"
    "net/http"
    "strings"
    "html/template"
    "github.com/SteelPangolin/gotoys/mac"
    "path"
    "archive/zip"
)

const browseBase = "/Volumes/media"

type Listing struct {
    BaseUrl        string
    Children    []os.FileInfo
}

func listDir(browsePath string) *Listing {
    listing := new(Listing)

    f, err := os.Open(browsePath)
    if (err != nil) {
        log.Panic("Couldn't open browse path: ", err)
    }
    defer f.Close()
    
    // TODO: filename encoding?
    listing.Children, err = f.Readdir(0) // all entries
    if (err != nil) {
        log.Panic("Couldn't list entire directory: ", err)
    }

    return listing
}

func pathDepth(p string) int {
    if p == "" {
        return 0
    }
    i := strings.Count(p, "/")
    if p[len(p)-1:] != "/" {
        i += 1
    }
    return i
}

func browseArchive(w http.ResponseWriter, zipPath string,
                   archiveRelPath string, baseUrl string,
                   action string) {
    z, err := zip.OpenReader(zipPath)
    if (err != nil) {
        log.Panic("Couldn't open browse path: ", err)
    }
    defer z.Close()

    var found *zip.File
    for _, f := range z.File {
        if f.Name == archiveRelPath {
            found = f
            break
        }
    }
    isDir := found == nil || found.FileInfo().IsDir()

    if isDir {
        listing := new(Listing)
        listing.BaseUrl = baseUrl
        
        var pred func(string) bool
        if archiveRelPath == "" {
            pred = func(p string) bool {
                log.Print("depth = ", pathDepth(p))
                return pathDepth(p) == 1
            }
        } else {
            archiveRelPathDepth := pathDepth(archiveRelPath)
            log.Print("archiveRelPathDepth = ", archiveRelPathDepth)
            pred = func(p string) bool {
                log.Print("depth = ", pathDepth(p))
                return strings.Index(p, archiveRelPath + "/") == 0 &&
                    pathDepth(p) == 1 + archiveRelPathDepth
            }
        }

        // TODO: Zip filename encoding?
        for _, f := range z.File {
            log.Print("member = ", f.Name)
            if pred(f.Name) {
                listing.Children = append(listing.Children, f.FileInfo())
            }
        }

        browseTemplate := template.Must(template.New("browseDir").
                                        Parse(browseDirHtml))
        err := browseTemplate.Execute(w, listing)
        if (err != nil) {
            log.Panic("Couldn't render template: ", err)
        }
    } else {
        if action == "download!" {

        } else {
            browseTemplate := template.Must(template.New("browseFile").
                                        Parse(browseFileHtml))
            filePage := FilePage{baseUrl, found}
            err := browseTemplate.Execute(w, filePage)
            if (err != nil) {
                log.Panic("Couldn't render template: ", err)
            }
        }
    }
}

const browseDirHtml = `<!DOCTYPE html>
<html>
    <head>
        <title>Mighty Browse</title>
    </head>
    <body>
        <table>
            <thead>
                <tr>
                    <th>Type</th>
                    <th>Name</th>
                    <th>Download</th>
                </tr>
            </thead>
            <tbody>
                {{ $baseurl := .BaseUrl }}
                {{ range .Children }}
                    <tr>
                        <td>
                            {{ if .IsDir }}
                                📁
                            {{ else }}
                                📄
                            {{ end }}
                        </td>
                        <td>
                            <a title='browse' href='{{ $baseurl }}/{{ .Name }}'>
                                {{ .Name }}
                            </a>
                        <td>
                        <td>
                            {{ if not .IsDir }}
                                <a title='download' href='{{ $baseurl }}/{{ .Name }}/download!'>
                                    ⬇︎
                                </a>
                            {{ end }}
                        <td>
                    </tr>
                {{ end }}
            </tbody>
        </table>
    </body>
</html>`

func browseDir(w http.ResponseWriter, browsePath string,
               baseUrl string) {
    listing := listDir(browsePath)
    listing.BaseUrl = baseUrl
    browseTemplate := template.Must(template.New("browseDir").
                                    Parse(browseDirHtml))
    err := browseTemplate.Execute(w, listing)
    if (err != nil) {
        log.Panic("Couldn't render template: ", err)
    }
}

const browseFileHtml = `<!DOCTYPE html>
<html>
    <head>
        <title>Mighty Browse</title>
    </head>
    <body>
        <a title='Download' href='{{ .BaseUrl }}/download!'>⬇︎</a>
        <code>{{ printf "%#q" .Info }}</code>
    </body>
</html>`

type FilePage struct {
    BaseUrl string
    Info    interface{}
}

func browseFile(w http.ResponseWriter, browsePath string, baseUrl string) {
    mddict := mac.Spotlight(browsePath)

    browseTemplate := template.Must(template.New("browseFile").
                                    Parse(browseFileHtml))
    filePage := FilePage{baseUrl, mddict}
    err := browseTemplate.Execute(w, filePage)
    if (err != nil) {
        log.Panic("Couldn't render template: ", err)
    }
}

const browseBaseUrl = "/"

func browse(w http.ResponseWriter, r *http.Request) {
    baseUrl := r.URL.Path

    baseUrlWithoutAction, action := path.Split(baseUrl)
    if action == "download!" {
        baseUrl = baseUrlWithoutAction
    } else {
        action = ""
    }

    browsePathParts := strings.Split(baseUrl[len(browseBaseUrl):], "archive!/")
    browsePath := path.Join(browseBase, browsePathParts[0])
    log.Print("browsePath: ", browsePath)

    var archiveRelPath string
    if len(browsePathParts) == 2 {
        archiveRelPath = browsePathParts[1]
    } else if len(browsePathParts) > 2 {
        log.Panic("TODO: Nested archives not supported yet")
    }

    fi, err := os.Stat(browsePath)
    if err != nil {
        log.Panic("Couldn't stat browse path: ", err)
    }
    mode := fi.Mode()
    if mode.IsRegular() {
            if path.Ext(browsePath) == ".zip" {
                if archiveRelPath == "" {
                    baseUrl = path.Join(r.URL.Path, "archive!")
                }
                browseArchive(w, browsePath, archiveRelPath, baseUrl, action)
            } else {
                if action == "download!" {
                    http.ServeFile(w, r, browsePath)
                } else {
                    browseFile(w, browsePath, baseUrl)
                }
            }
    } else if mode.IsDir() {
        browseDir(w, browsePath, baseUrl)
    } else {
        log.Panic("Unexpected file mode: ", mode.String())
    }
}

func main() {
    http.HandleFunc(browseBaseUrl, browse)
    http.ListenAndServe(":5000", nil)
}
