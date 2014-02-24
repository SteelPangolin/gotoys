package main

import (
    "github.com/SteelPangolin/gotoys/arclight"
    "fmt"
    "strings"
    "os"
)

func list(node arclight.Node, depth int) {
    pad := strings.Repeat("  ", depth)
    fmt.Printf("%s%s\n", pad, node.Name())
    dir, ok := node.(arclight.Dir)
    if ok {
        children, err := dir.Children()
        if err != nil {
            pad := strings.Repeat("  ", depth + 1)
            fmt.Printf("%s%#q\n", pad, err)
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
        fmt.Printf("%#q\n", err)
        return
    }
    root := arclight.NewOsNode(path, fi)
    list(root, 0)
}
