package main

import (
    "fmt"
    "flag"
    "github.com/SteelPangolin/gotoys/problems"
)

func main() {
    var err error

    flag.Parse()

    var filter problems.MessageFilter
    filter, err = problems.NewRegexpFilter([]string{
        // fragment
        `missing <!DOCTYPE> declaration`,
        `inserting implicit <body>`,
        `inserting missing 'title' element`,
        // Angular
        `proprietary attribute "ng-`,
    })
    if err != nil {
        fmt.Printf("Filter init error: %v\n", err)
        return
    }

    var tidy problems.Scanner
    tidy, err = problems.NewTidy()
    if err != nil {
        fmt.Printf("Tidy init error: %v\n", err)
        return
    }

    for _, path := range flag.Args() {
        msgs, err := tidy.ScanFile(filter, path)
        if err != nil {
            fmt.Printf("Tidy scan error for path %s: %v\n\n", path, err)
            continue
        }
        for _, msg := range msgs {
            fmt.Printf("%s:%v\n", path, msg)
        }
        fmt.Println("")
    }
}
