package main

import (
	"bytes"
	"compress/zlib"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
)

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

func main() {
	flag.Parse()
	path := flag.Arg(0)

	buf, err := ioutil.ReadFile(path)
	if err != nil {
		panic(err)
	}

	tokens, err := lex(buf)
	for _, token := range tokens {
		fmt.Printf("%s\n", token)
	}
	if err != nil {
		fmt.Printf("lexer error: %s\n", err)
	}

	streamIdx := 1
	for _, token := range tokens {
		if stream, ok := token.(*StreamToken); ok {
			streamPath := fmt.Sprintf("stream_%04d.dat", streamIdx)
			streamIdx++
			err := saveStream(stream, streamPath)
			if err != nil {
				fmt.Printf("error saving stream %s: %v\n", streamPath, err)
			}
		}
	}
}
