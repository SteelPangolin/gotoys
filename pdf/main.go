package main

import (
	"flag"
	"fmt"
	"io/ioutil"
)

/*
func saveStream(stream *StreamToken, path string) error {
	// TODO: check the filter type instead of assuming Flate
	buf := bytes.NewBuffer(stream.buf)
	r, err := zlib.NewReader(buf)
	if err != nil {
		return err
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = io.Copy(f, r)
	return err
}
*/

func main() {
	flag.Parse()
	path := flag.Arg(0)

	buf, err := ioutil.ReadFile(path)
	if err != nil {
		panic(err)
	}

	tokens, err := lex(buf)
	if err != nil {
		fmt.Printf("lexer error: %s\n", err)
		panic(err)
	}

	doc, stack, err := parse(tokens)
	fmt.Printf("doc: %v\n", doc)
	if err != nil || len(stack) > 0 {
		fmt.Printf("stack: %v\n", stack)
		fmt.Printf("parser error: %s\n", err)
	}
}
