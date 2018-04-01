package gollery

import (
	"io/ioutil"
	"path/filepath"
	"os"
	"io"
	"fmt"
	"log"
	"go/build"
)

func copyFile(src, dst string) (err error) {
	in, err := os.Open(src)
	if err != nil {
		return
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return
	}
	defer func() {
		if e := out.Close(); e != nil {
			err = e
		}
	}()

	_, err = io.Copy(out, in)
	if err != nil {
		return
	}

	err = out.Sync()
	if err != nil {
		return
	}

	si, err := os.Stat(src)
	if err != nil {
		return
	}
	err = os.Chmod(dst, si.Mode())
	if err != nil {
		return
	}

	return
}

func readDir(dir string) []os.FileInfo {
	files, err := ioutil.ReadDir(dir)
	check(err)

	return files
}

func removeFile(input string){
	if err := os.Remove(input); err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
	log.Println("File " + input + " successfull removed.")
}

func getDir() string {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	check(err)
	return dir
}

func getGoPath() string {
	gopath := os.Getenv("GOPATH")
	if gopath == "" {
		gopath = build.Default.GOPATH
	}
	return gopath
}


func check(e error) {
	if e != nil {
		log.Fatal(e)
	}
}

func checkFile (path string) bool {
	_, err := os.Stat(path)
	return os.IsNotExist(err)
}