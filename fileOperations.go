package gollery

import (
	"image"
	_ "image/jpeg" //read jpeg files
	_ "image/png"  //read png files
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/mholt/archiver"
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
	err := os.RemoveAll(filepath.FromSlash(input))
	check(err)

	log.Println("File " + input + " successfull removed.")
}

// Returns the current dir from where the application was started.
func getDir() string {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	check(err)
	return filepath.ToSlash(dir)
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
	defer f.Close()

	size, _, err := image.DecodeConfig(f)
	check(err)

	x, err := exif.Decode(f)
	var tm time.Time
	if err != nil {
		log.Print("Not able to read exif Data of " + path + ", please check! (" + err.Error() + ")")
		log.Print("Using current Date: " + time.Now().Format("Mon, 2 Jan 2006"))
		tm = time.Now()
	} else {
		tm, err = x.DateTime()
		check(err)
	}

	return tm.Format("Mon, 2 Jan 2006"), tm, float32(size.Width) / float32(size.Height)
}

// DownloadFile will download a url to a local file. It's efficient because it will
// write as it downloads and not load the whole file into memory.
func downloadFile(filepath string, url string) error {

	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Get the data
	log.Print("Downloading " + url + " - Please wait.")
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	return nil
}

// Creates a zip file at the output location with every given path within path[]
func addZip(c Config, gallery string) {
	if c.Galleries[gallery].Download {
		folders := []string{filepath.FromSlash(galleryPath + gallery + "/" + origImgDir), filepath.FromSlash(galleryPath + gallery + "/" + featImgDir)}
		err := archiver.Zip.Make(filepath.FromSlash(galleryPath+gallery+"/"+gallery+"_images.zip"), folders)
		check(err)
		log.Print("New zip file for " + gallery + " gallery created.")
	} else {
		if !checkFile(filepath.FromSlash(galleryPath + gallery + "/" + gallery + "_images.zip")) {
			removeFile(filepath.FromSlash(galleryPath + gallery + "/" + gallery + "_images.zip"))
		}
	}
}

func createCustomCss(c Config, gallery string) {
	if c.Galleries[gallery].CustomCss {
		if checkFile(galleryPath + "/" + gallery + "/custom_css") {
			file, err := os.Create(galleryPath + "/" + gallery + "/custom_css.css")
			check(err)
			log.Print("Created new file " + gallery + "/custom_css.css .")
			file.Close()
		} else {
			log.Print("Custom CSS file is already existing.")
		}
	} else {
		if !checkFile(filepath.FromSlash(galleryPath + gallery + "/custom_css.css")) {
			removeFile(filepath.FromSlash(galleryPath + gallery + "/custom_css.css"))
		}
	}
}
