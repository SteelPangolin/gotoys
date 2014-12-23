package main

import (
	"bufio"
	"encoding/binary"
	"flag"
	"fmt"
	"os"
)

import (
	"github.com/SteelPangolin/gotoys/arclight"
	"github.com/SteelPangolin/gotoys/dbm"
)

func findTextFiles(node arclight.VfsNode) {
	if dir, ok := node.(arclight.VfsDir); ok {
		children, err := dir.Children()
		if err != nil {
			fmt.Printf("dir listing error: %v\n", err)
			return
		}
		for _, child := range children {
			findTextFiles(child)
		}
	} else if mediatype, _ := node.MimeType(); mediatype == "text/plain" {
		if osFile, ok := node.(*arclight.OsFile); ok {
			fmt.Println(osFile.Path)
		}
		wordcount(node.(arclight.VfsFile))
	}
}

var Counts map[string]uint64 = make(map[string]uint64)
var Total uint64 = 0

func wordcount(node arclight.VfsFile) {
	reader, err := node.Open()
	if err != nil {
		fmt.Printf("open error: %v\n", err)
		return
	}
	defer reader.Close()

	scanner := bufio.NewScanner(reader)
	scanner.Split(bufio.ScanWords)
	for scanner.Scan() {
		word := scanner.Text()
		Counts[word]++
		Total++
	}
	if err := scanner.Err(); err != nil {
		fmt.Printf("scanner error: %v\n", err)
	}
}

func main() {
	flag.Parse()
	path := flag.Arg(0)

	fi, err := os.Stat(path)
	if err != nil {
		fmt.Printf("stat error: %v\n", err)
		return
	}

	root := arclight.NewOsNode(path, fi)
	findTextFiles(root)
	fmt.Printf("total: %v\n", Total)

	db, err := dbm.Open("wordcount")
	defer db.Close()
	if err != nil {
		fmt.Printf("DBM open error: %v\n", err)
		return
	}
	valueBuf := make([]byte, 8)
	for word, count := range Counts {
		binary.LittleEndian.PutUint64(valueBuf, count)
		err := db.Insert([]byte(word), valueBuf)
		if err != nil {
			fmt.Printf("DBM insert error: %v\n", err)
		}
	}
}
