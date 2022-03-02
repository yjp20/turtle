package main

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/yjp20/turtle/straw"
)

func main() {
	if len(os.Args) == 1 {
		println("Enter at least one or more files/directories to transform")
		return
	}

	for _, path := range os.Args[1:] {
		filepath.Walk(path, func (p string, i os.FileInfo, _ error) error {
			if i.IsDir() {
				return nil
			}

			if strings.HasSuffix(p, ".straw") {
				println(p)
				f, err := os.Open(p)
				if err != nil {
					println(err.Error())
				}
				b, err := ioutil.ReadAll(f)
				if err != nil {
					println(err.Error())
				}
				f, err = os.OpenFile(p, os.O_WRONLY, os.ModeAppend)
				if err != nil {
					println(err.Error())
				}
				_, err = f.Write(straw.Filter(b))
				if err != nil {
					println(err.Error())
				}
			}

			return nil
		})
	}
}
