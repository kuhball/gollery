package main

import (
	"log"
	"fmt"
	"os"
	"io/ioutil"
	"net/http"
	"html/template"
	"gopkg.in/yaml.v2"
	"strings"
	"path/filepath"
)

const origImgDir = "img/"
const prevImgDir = "preview/"
const featImgDir = "featured/"
const thumbImgDir = "thumbnail/"
const galleryPath = "static/gallery/"

const thumbSize = 396
const featSize = 796
const prevSize = 1080

var globConfig config

type config struct {
	Port      string
	Galleries map[string]*Galleries
}

type Galleries struct {
	Title       string
	Description string
	Download    bool
	Images      map[string]bool
	Dir 		dir
}

type dir struct {
	OrigImgDir  string
	PrevImgDir  string
	ThumbImgDir string
	FeatImgDir  string
	GalleryPath string
}

func initDir() dir {
	return dir{
		OrigImgDir:  origImgDir,
		PrevImgDir:  prevImgDir,
		ThumbImgDir: thumbImgDir,
		FeatImgDir:  featImgDir,
		GalleryPath: galleryPath,
	}
}

func readDir(dir string) []os.FileInfo {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		log.Fatal(err)
	}

	return files
}

func removeFile(input string) {
	if err := os.Remove(input); err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
	log.Println("File " + input + " successfull removed.")
}

func galleryHandler(w http.ResponseWriter, r *http.Request) {
	title := strings.Replace(r.URL.Path, "/", "", 2)

	if recreate {
		folders := []string{galleryPath + title + "/" + origImgDir, galleryPath + title + "/" + featImgDir}
		addZip(galleryPath+title+"/"+title+"_images.zip", folders)
		getFeatured(globConfig,title)
	}

	globConfig.Galleries[title].Dir = initDir()

	p, _ := globConfig.Galleries[title]
	t, _ := template.ParseFiles("gallery.html")
	t.Execute(w, p)
}

func staticHandler(w http.ResponseWriter, r *http.Request) {
	path := "." + r.URL.Path
	if f, err := os.Stat(path); err == nil && !f.IsDir() {
		http.ServeFile(w, r, path)
		return
	}

	http.NotFound(w, r)
}

func initWebServer(port string) {
	http.HandleFunc("/static/", staticHandler)

	for subSite := range globConfig.Galleries {
		http.HandleFunc("/"+subSite, galleryHandler)
	}

	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func readConfig(f string) config {
	if _, err := os.Stat(f); os.IsNotExist(err) {
		f, _ = filepath.Abs("./config.yaml")
	}

	var c config
	source, err := ioutil.ReadFile(f)
	if err != nil {
		panic(err)
	}

	err = yaml.Unmarshal(source, &c)
	if err != nil {
		panic(err)
	}
	for subSite := range c.Galleries {
		getFeatured(c, subSite)
	}

	return c
}

func getFeatured(c config, subSite string) config {
	c.Galleries[subSite].Images = make(map[string]bool)
	for _, orig := range readDir(galleryPath + subSite + "/" + origImgDir) {
		c.Galleries[subSite].Images[orig.Name()] = false
	}
	for _, orig := range readDir(galleryPath + subSite + "/" + featImgDir) {
		c.Galleries[subSite].Images[orig.Name()] = true
	}
	return c
}

func main() {
	globConfig = readConfig("")
	go initWebServer(globConfig.Port)
	checkSubSites(globConfig.Galleries)

	watchFile(globConfig.Galleries)
	//cliAccess()
}
