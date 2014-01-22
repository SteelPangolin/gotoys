package main

import (
	"flag"
	"fmt"
	"os"
	"container/list"
	"path/filepath"
)

func main() {
	flag.Parse()
	queue := list.New()
	for _, path := range flag.Args() {
		queue.PushBack(path)
	}
	numPathsLeft := queue.Len()

	totalPathsIndexed := 0

	// paths going to index workers
	pathsIn 	 := make(chan string)
	// more paths to index, coming back from index workers
	pathsOut     := make(chan string)
	// message that we've finished indexing a path
	pathIndexed  := make(chan bool)

	go indexWorker(pathsIn, pathsOut, pathIndexed)

	// hacky way to start for/select
	numPathsLeft++
	totalPathsIndexed--
	go startWorking(pathIndexed)
	
IndexMaster:
	for {
		select {
		case <-pathIndexed:
			numPathsLeft--
			totalPathsIndexed++
			fmt.Printf("path indexed: numPathsLeft = %d\n", numPathsLeft)
			front := queue.Front()
			if front == nil {
				if numPathsLeft == 0 {
					break IndexMaster
				}
			} else {
				fmt.Printf("sending path: numPathsLeft = %d\n", numPathsLeft)
				pathsIn <- queue.Remove(front).(string)
			}
		case path := <-pathsOut:
			numPathsLeft++
			queue.PushBack(path)
			fmt.Printf("subpath received: numPathsLeft = %d\n", numPathsLeft)
		}
	}

	fmt.Printf("Indexed %d paths.\n", totalPathsIndexed)
}

func startWorking(pathIndexed chan bool) {
	pathIndexed <- true
}

func indexWorker(pathsIn chan string, pathsOut chan string, pathIndexed chan bool) {
	for path := range pathsIn {
		index(path, pathsOut)
		pathIndexed <- true
	}
}

func index(path string, pathsOut chan string) {
	f, err := os.Open(path)
	if err != nil {
		fmt.Printf("Couldn't open %v: %v\n", path, err)
		return
	}
	defer f.Close()
	st, err := f.Stat()
	if err != nil {
		fmt.Printf("Couldn't stat %v: %v\n", path, err)
		return
	}
	if st.IsDir() {
		entries, err := f.Readdirnames(0 /* all of them */)
		if err != nil {
			fmt.Printf("Couldn't list directory %v: %v", path, err)
			return
		}
		fmt.Printf("%v is a directory with %d entries\n", path, len(entries))
		for _, name := range entries {
			pathsOut <- filepath.Join(path, name)
		}
	} else {
		fmt.Printf("%v is a file: %d bytes\n", path, st.Size());
	}
}