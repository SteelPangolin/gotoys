package main

import (
    "github.com/SteelPangolin/gotoys/arclight"
    "fmt"
    "strings"
    "os"
)

func list(node arclight.VfsNode, depth int) {
    pad := strings.Repeat("  ", depth)
    fmt.Printf("%s%s\n", pad, node.Name())

    pad = strings.Repeat("  ", depth + 1)

    for k, v := range node.Attrs() {
        fmt.Printf("%s%s: %#q\n", pad, k, v)
    }
    
    dir, ok := node.(arclight.VfsDir)
    if ok {
        children, err := dir.Children()
        if err != nil {
            fmt.Printf("%serror: %#q\n", pad, err)
            return
        }
        for _, child := range children {
            list(child, depth + 1)
        }
    }
}

func main() {
    path := "/Users/jehrhardt/Downloads"
    fi, err := os.Stat(path)
    if err != nil {
        fmt.Printf("error: %#q\n", err)
        return
    }
    root := arclight.NewOsNode(path, fi)
    list(root, 0)
}
