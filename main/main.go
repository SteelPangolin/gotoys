package main

import (
    "os"
    "log"
    "net/http"
    "encoding/json"
    "strings"
    "strconv"
    "html/template"
    "github.com/SteelPangolin/gotoys/mac"
    "path"
)

func echo(w http.ResponseWriter, r *http.Request) {
    var err error
    var reqJSON []byte
    reqJSON, err = json.MarshalIndent(r, "", "    ")
    if err != nil {
        log.Print(err)
        return
    }
    w.Header().Set("Content-Type", "application/json")
    _, err = w.Write(reqJSON)
    if err != nil {
        log.Print(err)
        return
    }
}

const maxBufLen = 4096
const upstream = "http://riverpig.lan"

func passthru(w http.ResponseWriter, r *http.Request) {
    log.Print(r.URL.Path)
    upstreamURL := strings.Join([]string{upstream, r.URL.Path}, "")
    resp, err := http.Get(upstreamURL)
    if err != nil {
        log.Print(err)
        return
    }
    defer resp.Body.Close()

    contentType := resp.Header.Get("Content-Type")
    if contentType != "" {
        w.Header().Set("Content-Type", contentType)
    }

    if resp.ContentLength != -1 {
        w.Header().Set("Content-Length", strconv.FormatInt(resp.ContentLength, 10))
    }

    w.WriteHeader(resp.StatusCode)

    buf := make([]byte, 4096)
    var n int
    for err = nil; err == nil; {
        n, err = resp.Body.Read(buf)
        w.Write(buf[:n])
    }
}

const browseBase = "/Users/jehrhardt/Downloads"

const browseDirHtml = `<!DOCTYPE html>
<html>
    <head>
        <title>Mighty Browse</title>
    </head>
    <body>
        <ul>
            {{ range . }}
                <li>
                    <a href='{{ . }}'>
                        {{ . }}
                    </a>
                </li>
            {{ end }}
        </ul>
    </body>
</html>`

func browseDir(w http.ResponseWriter, browsePath string) {
    f, err := os.Open(browsePath)
    defer f.Close()
    if (err != nil) {
        log.Panic("Couldn't open browse path: ", err)
    }
    
    names, err := f.Readdirnames(0) // all entries
    if (err != nil) {
        log.Panic("Couldn't list entire directory: ", err)
    }

    browseTemplate := template.Must(template.New("browseDir").
                                    Parse(browseDirHtml))
    err = browseTemplate.Execute(w, names)
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
        <code>{{ printf "%#q" . }}</code>
    </body>
</html>`

func browseFile(w http.ResponseWriter, browsePath string) {
    mddict := mac.Spotlight(browsePath)

    browseTemplate := template.Must(template.New("browseFile").
                                    Parse(browseFileHtml))
    err := browseTemplate.Execute(w, mddict)
    if (err != nil) {
        log.Panic("Couldn't render template: ", err)
    }
}

func browse(w http.ResponseWriter, r *http.Request) {
    browsePath := path.Join(browseBase, r.URL.Path[len("/browse/"):])
    log.Print("browsePath: ", browsePath)

    fi, err := os.Stat(browsePath)
    if err != nil {
        log.Panic("Couldn't stat browse path: ", err)
    }
    mode := fi.Mode()
    if mode.IsDir() {
        browseDir(w, browsePath)
    } else if mode.IsRegular() {
        browseFile(w, browsePath)
    } else {
        log.Panic("Unexpected file mode: ", mode.String())
    }
}

func main() {
    http.HandleFunc("/echo", echo)
    http.HandleFunc("/browse/", browse)
    http.HandleFunc("/", passthru)
    http.ListenAndServe(":5000", nil)
}
