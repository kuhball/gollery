package main

import (
	"log"
	"github.com/fsnotify/fsnotify"
	"regexp"
	"os/exec"
	"fmt"
	"os"
	"strconv"
)

const thumbSize = 384
const featSize = 514
const prevSize = 1080

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
		if (imgKind == origImgDir) {

		} else if (imgKind == featImgDir) {
			createImage(event.Name, galleryPath+subSite+thumbImgDir+"thumb"+filename, featSize)
		}
		createImage(event.Name, galleryPath+subSite+prevImgDir+"prev"+filename, prevSize)
	} else if event.Op.String() == "REMOVE" {
		removeFile(galleryPath + subSite + prevImgDir + "prev" + filename)
		removeFile(galleryPath + subSite + thumbImgDir + "thumb" + filename)
	}
}

//func createThumb(input string, output string) {
//	cmd := "magick"
//	args := []string{input, "-define", "jpeg:size=786x", "-auto-orient", "-quality", "80", "-thumbnail", "384x", "-unsharp", "0x.5", output}
//	if err := exec.Command(cmd, args...).Run(); err != nil {
//		fmt.Fprintln(os.Stderr, err)
//		os.Exit(1)
//	}
//	fmt.Println("Successfully created thumbnail.")
//}

func createImage(input string, output string, size int) {
	cmd := "magick"
	args := []string{input, "-define", "jpeg:size=" + "x" + strconv.Itoa(size*2), "-auto-orient", "-quality", "80", "-thumbnail", "x" + strconv.Itoa(size), "-unsharp", "0x.5", output}
	log.Println(args)
	if err := exec.Command(cmd, args...).Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println("Successfully created " + strconv.Itoa(size))
}

//func createPrev(input string, output string) {
//	cmd := "magick"
//	args := []string{input, "-define", "jpeg:size=2160x", "-auto-orient", "-thumbnail", "1080x", "-unsharp", "0x.5", output}
//	if err := exec.Command(cmd, args...).Run(); err != nil {
//		fmt.Fprintln(os.Stderr, err)
//		os.Exit(1)
//	}
//	fmt.Println("Successfully created preview.")
//}

func checkSubSites(subSites []string) {
	for _, subSite := range subSites {
		checkFiles(readDir(galleryPath+subSite+origImgDir), subSite, false)
		checkFiles(readDir(galleryPath+subSite+featImgDir), subSite, true)
	}
}

func checkFiles(files []os.FileInfo, subSite string, featured bool) {
	for _, file := range files {
		if _, err := os.Stat(galleryPath + subSite + thumbImgDir + "thumb" + file.Name()); os.IsNotExist(err) {
			if featured {
				createImage(galleryPath+subSite+featImgDir+file.Name(), galleryPath+subSite+thumbImgDir+"thumb"+file.Name(), featSize)
			} else {
				createImage(galleryPath+subSite+origImgDir+file.Name(), galleryPath+subSite+thumbImgDir+"thumb"+file.Name(), thumbSize)
			}
		}
		if _, err := os.Stat(galleryPath + subSite + prevImgDir + "prev" + file.Name()); os.IsNotExist(err) {
			if featured {
				createImage(galleryPath+subSite+featImgDir+file.Name(), galleryPath+subSite+prevImgDir+"prev"+file.Name(), prevSize)

			} else {
				createImage(galleryPath+subSite+origImgDir+file.Name(), galleryPath+subSite+prevImgDir+"prev"+file.Name(), prevSize)
			}
		}
	}
}

//func checkFiles(subSites []string) {
//	for _, file := range files {
//		if _, err := os.Stat(galleryPath + path + thumbImgDir + "thumb" + file.Name()); os.IsNotExist(err) {
//			createThumb(galleryPath+path+origImgDir+file.Name(), galleryPath+path+thumbImgDir+"thumb"+file.Name())
//		}
//		if _, err := os.Stat(galleryPath + path + prevImgDir + "prev" + file.Name()); os.IsNotExist(err) {
//			createImage(galleryPath+path+origImgDir+file.Name(), galleryPath+path+prevImgDir+"prev"+file.Name(), prevSize)
//		}
//	}
//}
