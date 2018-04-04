package gollery

import (
	"io/ioutil"
	"path/filepath"
	"os"
	"fmt"
	"log"
	"go/build"

	"image"
	"github.com/xor-gate/goexif2/exif"
	 _ "image/jpeg"
	 _ "image/png"
	"time"
)

func readDir(dir string) []os.FileInfo {
	files, err := ioutil.ReadDir(filepath.FromSlash(dir))
	check(err)

	return files
}

func removeFile(input string) {
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

func checkFile(path string) bool {
	_, err := os.Stat(filepath.FromSlash(path))
	return os.IsNotExist(err)
}

func returnImageData(path string) (string, time.Time, float32) {
	f, err := os.Open(filepath.FromSlash(path))
	check(err)

	size, _, err := image.DecodeConfig(f)
	check(err)

	x, err := exif.Decode(f)
	check(err)

	tm, err := x.DateTime()
	check(err)

	return tm.Format("Mon, 2 Jan 2006"), tm ,float32(size.Width) / float32(size.Height)
}
