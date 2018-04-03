package gollery

import (
	"log"
	"github.com/fsnotify/fsnotify"
	"regexp"
	"os/exec"
	"fmt"
	"os"
	"strconv"
	"github.com/mholt/archiver"
	"strings"
)

var recreate = false

func watchFile(subSites map[string]*Galleries) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	done := make(chan bool)
	go func() {
		for {
			select {
			case event := <-watcher.Events:
				log.Println("event:", event)
				filterFile(event)
				recreate = true
			case err := <-watcher.Errors:
				log.Println("error:", err)
			}
		}
	}()

	for subSite := range subSites {
		err = watcher.Add(galleryPath + subSite + "/" + origImgDir)
		err = watcher.Add(galleryPath + subSite + "/" + featImgDir)
		if err != nil {
			log.Fatal(err)
		}
	}
	<-done
}

func filterFile(event fsnotify.Event) {
	filenameReplace, _ := regexp.Compile(`(\w:)*(\w*\\)`)
	filename := filenameReplace.ReplaceAllString(event.Name, "")

	subSiteReplace, _ := regexp.Compile(`(\w*)`)
	subSite := subSiteReplace.FindString(event.Name[len(galleryPath):]) + "/"

	imgKindReplace, _ := regexp.Compile(`(\w*)`)
	imgKind := imgKindReplace.FindString(event.Name[len(galleryPath+subSite):]) + "/"

	if event.Op.String() == "CREATE" {
		if imgKind == origImgDir {
			go createImage(event.Name, galleryPath+subSite+thumbImgDir+"thumb"+filename, thumbSize)
		} else if imgKind == featImgDir {
			go createImage(event.Name, galleryPath+subSite+thumbImgDir+"thumb"+filename, featSize)
		}
		go createImage(event.Name, galleryPath+subSite+prevImgDir+"prev"+filename, prevSize)
	} else if event.Op.String() == "REMOVE" {
		log.Print(filename)
		removeFile(galleryPath + subSite + prevImgDir + "prev" + filename)
		removeFile(galleryPath + subSite + thumbImgDir + "thumb" + filename)
	}
}

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

func checkSubSites(subSites map[string]*Galleries) {
	for subSite := range subSites {
		subSite = subSite + "/"
		checkFiles(readDir(galleryPath+subSite+origImgDir), subSite, false)
		checkFiles(readDir(galleryPath+subSite+featImgDir), subSite, true)

		folders := []string{galleryPath + subSite + origImgDir, galleryPath + subSite + featImgDir}
		addZip(galleryPath+subSite+strings.Replace(subSite, "/", "", 1)+"_images.zip", folders)
	}
}

func checkFiles(files []os.FileInfo, subSite string, featured bool) {
	for _, file := range files {
		if checkFile(galleryPath + subSite + thumbImgDir + "thumb" + file.Name()) {
			if featured {
				go createImage(galleryPath+subSite+featImgDir+file.Name(), galleryPath+subSite+thumbImgDir+"thumb"+file.Name(), featSize)
			} else {
				go createImage(galleryPath+subSite+origImgDir+file.Name(), galleryPath+subSite+thumbImgDir+"thumb"+file.Name(), thumbSize)
			}
		}
		if checkFile(galleryPath + subSite + prevImgDir + "prev" + file.Name()) {
			if featured {
				go createImage(galleryPath+subSite+featImgDir+file.Name(), galleryPath+subSite+prevImgDir+"prev"+file.Name(), prevSize)

			} else {
				go createImage(galleryPath+subSite+origImgDir+file.Name(), galleryPath+subSite+prevImgDir+"prev"+file.Name(), prevSize)
			}
		}
	}
}

func addZip(output string, path []string) {
	err := archiver.Zip.Make(output, path)
	if err != nil {
		log.Fatal(err)
	}
}
