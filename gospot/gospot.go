package main

import (
	"encoding/json"
	"flag"
	"os"
)

import (
	"github.com/SteelPangolin/gotoys/mac"
)

func main() {
	flag.Parse()
	for _, path := range flag.Args() {
		attrs, err := mac.Spotlight(path)
		if err != nil {
			os.Stderr.WriteString(err.Error())
			continue
		}
		indentedJson, err := json.MarshalIndent(attrs, "", "    ")
		if err != nil {
			os.Stderr.WriteString(err.Error())
			continue
		}
		os.Stdout.Write(indentedJson)
		os.Stdout.WriteString("\n")
	}
}
