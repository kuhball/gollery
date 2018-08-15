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
	"runtime"
	"sync"
	"time"
)

var (
	cmd             string
	recreate        = false
	watcher         *fsnotify.Watcher
	configWriteTime = time.Now()
	wg              sync.WaitGroup
	semaphore       = make(chan struct{}, runtime.NumCPU())
)

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
				go filterFile(event)
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
	filenameReplace, _ := regexp.Compile(`(\w:)*(\w*\\)|(/.*/)`)
	filename := filenameReplace.ReplaceAllString(event.Name, "")

	galleryReplace, _ := regexp.Compile(`(\w*)`)
	gallery := galleryReplace.FindString(event.Name[len(galleryPath):]) + "/"

	imgKind := galleryReplace.FindString(event.Name[len(galleryPath+gallery):]) + "/"

	if event.Op.String() == "CREATE" {
		if imgKind == origImgDir {
			wg.Add(1)
			createImage(event.Name, galleryPath+gallery+thumbImgDir+"thumb"+filename, thumbSize)
		} else if imgKind == featImgDir {
			wg.Add(1)
			createImage(event.Name, galleryPath+gallery+thumbImgDir+"feat"+filename, featSize)
		}
		wg.Add(1)
		createImage(event.Name, galleryPath+gallery+prevImgDir+"prev"+filename, prevSize)
		recreate = true
		wg.Wait()
	} else if event.Op.String() == "WRITE" && filename == "config.yaml" {
		if duration := time.Since(configWriteTime); duration.Seconds() > time.Second.Seconds()*3 {
			log.Print("Reading changes from config.yaml.")
			newConfig := ReadConfig(configPath, true)
			for key, gallery := range newConfig.Galleries {
				if gallery != GlobConfig.Galleries[gallery.Title] {
					go func() {
						time.Sleep(3 * time.Second)
						err := watcher.Add(galleryPath + gallery.Title + "/" + origImgDir)
						check(err)
						time.Sleep(3 * time.Second)
						err = watcher.Add(galleryPath + gallery.Title + "/" + featImgDir)
						check(err)
					}()
					if GlobConfig.Galleries[key] == nil || gallery.Title != GlobConfig.Galleries[key].Title {
						createGalleryHandle(*gallery)
					}
					if GlobConfig.Galleries[key] != nil && gallery.Download != GlobConfig.Galleries[key].Download {
						addZip(newConfig, key)
					}
					createCustomCss(newConfig, key)
				}
				gallery.Dir = initDir()
			}
			GlobConfig = newConfig
			configWriteTime = time.Now()
		}
	} else if event.Op.String() == "REMOVE" || event.Op.String() == "RENAME" {
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
	go func() {
		semaphore <- struct{}{} // Lock
		defer func() {
			<-semaphore // Unlock
			wg.Done()
		}()
		args := []string{input, "-define", "jpeg:size=" + strconv.Itoa(size*2) + "x", "-auto-orient", "-quality", "80", "-thumbnail", strconv.Itoa(size) + "x", "-unsharp", "0x.5", output}
		if err := exec.Command(cmd, args...).Run(); err != nil {
			fmt.Fprintln(os.Stderr, err, input)
			os.Exit(1)
		}
		fmt.Println("Successfully created " + strconv.Itoa(size) + " - " + output)
	}()
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
			wg.Add(1)
			createImage(galleryPath+gallery+origImgDir+file.Name(), galleryPath+gallery+thumbImgDir+"thumb"+file.Name(), thumbSize)
		} else if checkFile(galleryPath+gallery+thumbImgDir+"feat"+file.Name()) && featured {
			wg.Add(1)
			createImage(galleryPath+gallery+featImgDir+file.Name(), galleryPath+gallery+thumbImgDir+"feat"+file.Name(), featSize)
		}
		if checkFile(galleryPath + gallery + prevImgDir + "prev" + file.Name()) {
			if featured {
				wg.Add(1)
				createImage(galleryPath+gallery+featImgDir+file.Name(), galleryPath+gallery+prevImgDir+"prev"+file.Name(), prevSize)
			} else {
				wg.Add(1)
				createImage(galleryPath+gallery+origImgDir+file.Name(), galleryPath+gallery+prevImgDir+"prev"+file.Name(), prevSize)
			}
		}
	}
	wg.Wait()
	if featured {
		log.Print("Featured thumbnails and previews for " + gallery + " are available.")
	} else {
		log.Print("Normal thumbnails and previews for " + gallery + " are available.")
	}

}
