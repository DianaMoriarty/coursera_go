package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
)

func dirTreePrint(out io.Writer, path string, printFiles bool, indent string) error {
	directory, err := os.Open(path)
	defer directory.Close()
	if err != nil {
		return err
	}

	files, err := directory.Readdir(-1)
	if err != nil {
		return err
	}

	if !printFiles {
		var dirs []os.FileInfo
		for _, file := range files {
			if file.IsDir() {
				dirs = append(dirs, file)
			}
		}
		files = dirs
	}

	sort.Slice(files, func(i, j int) bool {
		return files[i].Name() < files[j].Name()
	})

	var symbol, next string
	for idx, file := range files {
		if idx == len(files)-1 {
			symbol = "└───"
			next = indent + "\t"
		} else {
			symbol = "├───"
			next = indent + "│\t"
		}

		if file.IsDir() {
			fmt.Fprintf(out, "%s%s%s\n", indent, symbol, file.Name())
			err = dirTreePrint(out, filepath.Join(path, file.Name()), printFiles, next)
			if err != nil {
				return err
			}
		} else {
			var fileSize string
			if file.Size() == 0 {
				fileSize = "empty"
			} else {
				fileSize = fmt.Sprintf("%db", file.Size())
			}
			fmt.Fprintf(out, "%s%s%s (%s)\n", indent, symbol, file.Name(), fileSize)
		}
	}
	return nil
}

func dirTree(out io.Writer, path string, printFiles bool) error {
	return dirTreePrint(out, path, printFiles, "")
}

func main() {
	out := os.Stdout
	if !(len(os.Args) == 2 || len(os.Args) == 3) {
		panic("usage go run main.go . [-f]")
	}
	path := os.Args[1]
	printFiles := len(os.Args) == 3 && os.Args[2] == "-f"
	err := dirTree(out, path, printFiles)
	if err != nil {
		panic(err.Error())
	}
}
