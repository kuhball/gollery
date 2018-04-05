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
	"github.com/mholt/archiver"
)

// used for zip recreation - changed by watcher event
var recreate = false

// initialize a new fsnotify watcher
// watcher calls filterfile() in case of an event
// the origImgDir and featImgDir from all galleries are added to the watcher
func watchFile(galleries map[string]*Gallery) {
	watcher, err := fsnotify.NewWatcher()
	check(err)

	defer watcher.Close()

	done := make(chan bool)
	go func() {
		for {
			select {
			case event := <-watcher.Events:
				log.Println("event:", event)
				filterFile(event)
				recreate = true
			case err = <-watcher.Errors:
				log.Println("error:", err)
			}
		}
	}()

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
			go createImage(event.Name, galleryPath+gallery+thumbImgDir+"thumb"+filename, thumbSize)
		} else if imgKind == featImgDir {
			go createImage(event.Name, galleryPath+gallery+thumbImgDir+"feat"+filename, featSize)
		}
		go createImage(event.Name, galleryPath+gallery+prevImgDir+"prev"+filename, prevSize)
	} else if event.Op.String() == "REMOVE" {
		log.Print(filename)
		removeFile(galleryPath + gallery + prevImgDir + "prev" + filename)
		if imgKind == origImgDir {
			removeFile(galleryPath + gallery + thumbImgDir + "thumb" + filename)
		} else if imgKind == featImgDir {
			removeFile(galleryPath + gallery + thumbImgDir + "feat" + filename)
		}
	}
}

// This function calls the cli tool magick
// The input image is loaded and a new output image is created with a given size
// Magick options are adjusted for thumbnails.
func createImage(input string, output string, size int) {
	cmd := "magick"
	args := []string{input, "-define", "jpeg:size=" + strconv.Itoa(size*2) + "x", "-auto-orient", "-quality", "80", "-thumbnail", strconv.Itoa(size) + "x", "-unsharp", "0x.5", output}
	log.Println(args)
	if err := exec.Command(cmd, args...).Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println("Successfully created " + strconv.Itoa(size))
}

// Calls checkFiles for every galleries orig and feat images
// Creates a zip file with all images.
// Called at the start of gollery.
func checkSubSites(galleries map[string]*Gallery) {
	for gallery := range galleries {
		gallery = gallery + "/"
		checkFiles(readDir(galleryPath+gallery+origImgDir), gallery, false)
		checkFiles(readDir(galleryPath+gallery+featImgDir), gallery, true)

		folders := []string{galleryPath + gallery + origImgDir, galleryPath + gallery + featImgDir}
		addZip(galleryPath+gallery+strings.Replace(gallery, "/", "", 1)+"_images.zip", folders)
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
}

// Creates a zip file at the output location with every given path within path[]
func addZip(output string, path []string) {
	err := archiver.Zip.Make(output, path)
	check(err)
}
