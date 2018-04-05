package gollery

import (
	"go/build"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"image"
	_ "image/jpeg" //read jpeg files
	_ "image/png"  //read png files
	"time"

	"github.com/xor-gate/goexif2/exif"
)

// Returns the content of a directory on a filesystem.
func readDir(dir string) []os.FileInfo {
	files, err := ioutil.ReadDir(filepath.FromSlash(dir))
	check(err)

	return files
}

// Delete a file from a filesystem.
func removeFile(input string) {
	err := os.Remove(filepath.FromSlash(input))
	check(err)

	log.Println("File " + input + " successfull removed.")
}

// Returns the current dir from where the application was started.
func getDir() string {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	check(err)
	return filepath.ToSlash(dir)
}

// Returns the system GO Path.
func getGoPath() string {
	gopath := os.Getenv("GOPATH")
	if gopath == "" {
		gopath = build.Default.GOPATH
	}
	return filepath.ToSlash(gopath)
}

// Simple error check and log in case of != nil.
func check(e error) {
	if e != nil {
		log.Fatal(e)
	}
}

// Checks whether file is existing on a filesystem and returns true if not existing.
func checkFile(path string) bool {
	_, err := os.Stat(filepath.FromSlash(path))
	return os.IsNotExist(err)
}

// This functions reads an image from a filesystem.
// It decodes the image size and the taken time/date from the image.
// Returns the Date as a string, the time object and the Image ratio (width/height).
func returnImageData(path string) (string, time.Time, float32) {
	f, err := os.Open(filepath.FromSlash(path))
	check(err)

	size, _, err := image.DecodeConfig(f)
	check(err)

	x, err := exif.Decode(f)
	check(err)

	tm, err := x.DateTime()
	check(err)

	return tm.Format("Mon, 2 Jan 2006"), tm, float32(size.Width) / float32(size.Height)
}
