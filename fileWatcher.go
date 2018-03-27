package main

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

const thumbSize = 396
const featSize = 796
const prevSize = 1080

var recreateZip = false

func watchFile(subSites [] string) {
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
				recreateZip = true
			case err := <-watcher.Errors:
				log.Println("error:", err)
			}
		}
	}()

	for _, subSite := range subSites {
		err = watcher.Add(galleryPath + subSite + origImgDir)
		err = watcher.Add(galleryPath + subSite + featImgDir)
		if err != nil {
			log.Fatal(err)
		}
	}
	<-done
}

func filterFile(event fsnotify.Event) {
	filenameReplace, _ := regexp.Compile(`(\w*\\)`)
	filename := filenameReplace.ReplaceAllString(event.Name, "")

	subSiteReplace, _ := regexp.Compile(`(\w*)`)
	subSite := subSiteReplace.FindString(event.Name[len(galleryPath):]) + "/"

	imgKindReplace, _ := regexp.Compile(`(\w*)`)
	imgKind := imgKindReplace.FindString(event.Name[len(galleryPath+subSite):]) + "/"

	if event.Op.String() == "CREATE" {
		if imgKind == origImgDir {

		} else if imgKind == featImgDir {
			createImage(event.Name, galleryPath+subSite+thumbImgDir+"thumb"+filename, featSize)
		}
		createImage(event.Name, galleryPath+subSite+prevImgDir+"prev"+filename, prevSize)
	} else if event.Op.String() == "REMOVE" {
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

func checkSubSites(subSites []string) {
	for _, subSite := range subSites {
		checkFiles(readDir(galleryPath+subSite+origImgDir), subSite, false)
		checkFiles(readDir(galleryPath+subSite+featImgDir), subSite, true)

		folders := []string{galleryPath + subSite + origImgDir, galleryPath + subSite + featImgDir}
		addZip(galleryPath+subSite+strings.Replace(subSite, "/", "", 1)+"_images.zip", folders)
	}
}

func checkFiles(files []os.FileInfo, subSite string, featured bool) {
	for _, file := range files {
		if _, err := os.Stat(galleryPath + subSite + thumbImgDir + "thumb" + file.Name()); os.IsNotExist(err) {
			if featured {
				go createImage(galleryPath+subSite+featImgDir+file.Name(), galleryPath+subSite+thumbImgDir+"thumb"+file.Name(), featSize)
			} else {
				go createImage(galleryPath+subSite+origImgDir+file.Name(), galleryPath+subSite+thumbImgDir+"thumb"+file.Name(), thumbSize)
			}
		}
		if _, err := os.Stat(galleryPath + subSite + prevImgDir + "prev" + file.Name()); os.IsNotExist(err) {
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
