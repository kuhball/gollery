package gollery

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"github.com/fsnotify/fsnotify"
	"time"
)

// variable for image creation tool convert / magick
var cmd string

// used for zip recreation - changed by watcher event
var recreate = false

var watcher *fsnotify.Watcher

var configWriteTime = time.Now()

// initialize a new fsnotify watcher
// watcher calls filterfile() in case of an event
// the origImgDir and featImgDir from all galleries are added to the watcher
func watchFile(galleries map[string]*Gallery) {
	var err error
	watcher, err = fsnotify.NewWatcher()
	check(err)
	defer watcher.Close()

	done := make(chan bool)
	go func() {
		for {
			select {
			case event := <-watcher.Events:
				filterFile(event)
			case err = <-watcher.Errors:
				log.Println("error:", err)
			}
		}
	}()

	err = watcher.Add(galleryPath + "config.yaml")
	check(err)

	for gallery := range galleries {
		err = watcher.Add(galleryPath + gallery + "/" + origImgDir)
		check(err)
		err = watcher.Add(galleryPath + gallery + "/" + featImgDir)
		check(err)
	}
	<-done
}

// regex expressions for getting the filename, gallery & image kind (orig / feat)
// In case of a "CREATE" operation preview and thumbnail images are created with different sizes for feat and orig.
// In case of a "REMOVE" operation preview and thumbnail images are removed from the filesystem
func filterFile(event fsnotify.Event) {
	filenameReplace, _ := regexp.Compile(`(\w:)*(\w*\\)`)
	filename := filenameReplace.ReplaceAllString(event.Name, "")

	galleryReplace, _ := regexp.Compile(`(\w*)`)
	gallery := galleryReplace.FindString(event.Name[len(galleryPath):]) + "/"

	imgKindReplace, _ := regexp.Compile(`(\w*)`)
	imgKind := imgKindReplace.FindString(event.Name[len(galleryPath+gallery):]) + "/"

	if event.Op.String() == "CREATE" {
		if imgKind == origImgDir {
			createImage(event.Name, galleryPath+gallery+thumbImgDir+"thumb"+filename, thumbSize)
			GlobConfig = appendImage(GlobConfig, gallery[:len(gallery)-1], filename, false)
		} else if imgKind == featImgDir {
			createImage(event.Name, galleryPath+gallery+thumbImgDir+"feat"+filename, featSize)
			GlobConfig = appendImage(GlobConfig, gallery[:len(gallery)-1], filename, true)
		}
		go createImage(event.Name, galleryPath+gallery+prevImgDir+"prev"+filename, prevSize)
		GlobConfig = sortImages(GlobConfig, gallery[:len(gallery)-1])
		recreate = true
	} else if event.Op.String() == "WRITE" && filename == "config.yaml" {
		if duration := time.Since(configWriteTime); duration.Seconds() > time.Second.Seconds()*3 {
			log.Print("Reading changes from config.yaml.")
			newConfig := ReadConfig(configPath, true)
			for _, gallery := range newConfig.Galleries {
				if gallery != GlobConfig.Galleries[gallery.Title] {
					go func() {
						time.Sleep(3 * time.Second)
						err := watcher.Add(galleryPath + gallery.Title + "/" + origImgDir)
						check(err)
						time.Sleep(3 * time.Second)
						err = watcher.Add(galleryPath + gallery.Title + "/" + featImgDir)
						check(err)
					}()
					if GlobConfig.Galleries[gallery.Title] == nil || gallery.Title != GlobConfig.Galleries[gallery.Title].Title {
						createGalleryHandle(newConfig, gallery.Title)
					}
					addZip(newConfig, gallery.Title)
					createCustomCss(newConfig, gallery.Title)
				}
				gallery.Dir = initDir()
			}
			GlobConfig = newConfig
			configWriteTime = time.Now()
		}
	} else if event.Op.String() == "REMOVE" {
		log.Print("Removing Image " + filename)
		removeFile(galleryPath + gallery + prevImgDir + "prev" + filename)
		if imgKind == origImgDir {
			removeFile(galleryPath + gallery + thumbImgDir + "thumb" + filename)
		} else if imgKind == featImgDir {
			removeFile(galleryPath + gallery + thumbImgDir + "feat" + filename)
		}
		GlobConfig.Galleries[gallery[:len(gallery)-1]] = deleteImage(GlobConfig.Galleries[gallery[:len(gallery)-1]], filename)
		recreate = true
	}
}

// Check whether magick or convert are available for exec & print version
func checkImageTool() {
	path, err := exec.LookPath("convert")
	if err != nil {
		log.Print("convert is not available.")
	}
	if strings.Contains(path, "system32") {
		log.Print("convert is not available.")
	} else {
		cmd = "convert"
		log.Print("convert it is.")
		return
	}

	path, err = exec.LookPath("magick")
	if err != nil {
		log.Fatal("installing fortune is in your future")
	} else {
		cmd = "magick"
		log.Print("magick it is.")
		return
	}

	log.Fatal("Convert and magick are not available, please fix that!")
}

// This function calls the cli tool magick / convert
// The input image is loaded and a new output image is created with a given size. An error is returned in case of a not valid image.
// Magick options are adjusted for thumbnails.
func createImage(input string, output string, size int) {
	args := []string{input, "-define", "jpeg:size=" + strconv.Itoa(size*2) + "x", "-auto-orient", "-quality", "80", "-thumbnail", strconv.Itoa(size) + "x", "-unsharp", "0x.5", output}
	if err := exec.Command(cmd, args...).Run(); err != nil {
		fmt.Fprintln(os.Stderr, err, input)
		os.Exit(1)
	}
	fmt.Println("Successfully created " + strconv.Itoa(size) + " - " + output)
}

// Calls checkFiles for every galleries orig and feat images
// Creates a zip file with all images.
// Called at the start of gollery.
func checkSubSites(galleries map[string]*Gallery) {
	for gallery := range galleries {
		gallery = gallery + "/"
		checkFiles(readDir(galleryPath+gallery+origImgDir), gallery, false)
		checkFiles(readDir(galleryPath+gallery+featImgDir), gallery, true)

		addZip(GlobConfig, strings.Replace(gallery, "/", "", 1))
	}
}

// Checks whether a thumbnail and preview has been created for all given images.
// Creates thumbnail & preview in case they are not existing.
func checkFiles(files []os.FileInfo, gallery string, featured bool) {
	for _, file := range files {
		if checkFile(galleryPath+gallery+thumbImgDir+"thumb"+file.Name()) && !featured {
			go createImage(galleryPath+gallery+origImgDir+file.Name(), galleryPath+gallery+thumbImgDir+"thumb"+file.Name(), thumbSize)
		} else if checkFile(galleryPath+gallery+thumbImgDir+"feat"+file.Name()) && featured {
			go createImage(galleryPath+gallery+featImgDir+file.Name(), galleryPath+gallery+thumbImgDir+"feat"+file.Name(), featSize)
		}
		if checkFile(galleryPath + gallery + prevImgDir + "prev" + file.Name()) {
			if featured {
				go createImage(galleryPath+gallery+featImgDir+file.Name(), galleryPath+gallery+prevImgDir+"prev"+file.Name(), prevSize)
			} else {
				go createImage(galleryPath+gallery+origImgDir+file.Name(), galleryPath+gallery+prevImgDir+"prev"+file.Name(), prevSize)
			}
		}
	}
	if featured {
		log.Print("Featured thumbnails and previews for " + gallery + " are available.")
	} else {
		log.Print("Normal thumbnails and previews for " + gallery + " are available.")
	}

}
