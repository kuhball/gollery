package gollery

import (
	"io/ioutil"
	"path/filepath"
	"os"
	"fmt"
	"log"
	"go/build"
)

func readDir(dir string) []os.FileInfo {
	files, err := ioutil.ReadDir(filepath.FromSlash(dir))
	check(err)

	return files
}

func removeFile(input string){
	if err := os.Remove(filepath.FromSlash(input)); err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
	log.Println("File " + input + " successfull removed.")
}

func getDir() string {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	check(err)
	return filepath.ToSlash(dir)
}

func getGoPath() string {
	gopath := os.Getenv("GOPATH")
	if gopath == "" {
		gopath = build.Default.GOPATH
	}
	return filepath.ToSlash(gopath)
}


func check(e error) {
	if e != nil {
		log.Fatal(e)
	}
}

func checkFile (path string) bool {
	_, err := os.Stat(filepath.FromSlash(path))
	return os.IsNotExist(err)
}