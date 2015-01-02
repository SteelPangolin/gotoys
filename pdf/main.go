package main

import (
	"flag"
	"fmt"
	"io/ioutil"
)

func mustParse(buf []byte) Document {
	tokens, err := lex(buf)
	if err != nil {
		fmt.Printf("lexer error: %s\n", err)
		panic(err)
	}

	doc, err := parse(tokens)
	if err != nil {
		fmt.Printf("parser error: %s\n", err)
		panic(err)
	}

	return doc
}

func findPages(node PDFMap) []PDFMap {
	if node["Type"].Val().(string) == "Page" {
		return []PDFMap{node}
	}

	pages := []PDFMap{}
	for _, kidRef := range node["Kids"].Val().(PDFList) {
		for _, page := range findPages(kidRef.Val().(PDFMap)) {
			pages = append(pages, page)
		}
	}
	return pages
}

func getContents(page PDFMap) [][]byte {
	defer func() {
		if e := recover(); e != nil {
			fmt.Printf("stream error: %v\n", e)
		}
	}()
	switch contents := page["Contents"].Val().(type) {
	case []byte:
		return [][]byte{contents}
	case PDFList:
		// should concatenate, but also want to know why multiple content streams
		contentsList := [][]byte{}
		for _, contentsRef := range contents {
			contentsList = append(contentsList, contentsRef.Val().([]byte))
		}
		return contentsList
	default:
		panic(fmt.Errorf("Page: illegal contents: %T %s", contents, contents))
	}
}

func main() {
	flag.Parse()
	path := flag.Arg(0)

	buf, err := ioutil.ReadFile(path)
	if err != nil {
		panic(err)
	}

	doc := mustParse(buf)
	for _, trailer := range doc.Trailers {
		if rootRef, hasRoot := trailer["Root"]; hasRoot {
			root := rootRef.Val().(PDFMap)
			pages := findPages(root["Pages"].Val().(PDFMap))
			for _, page := range pages {
				for _, contents := range getContents(page) {
					func() {
						fmt.Printf("%q\n", contents)
						return
					}()
				}
			}
		}
	}
}
