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
)

const origImgDir = "img/"
const prevImgDir = "preview/"
const featImgDir = "featured/"
const thumbImgDir = "thumbnail/"
const galleryPath = "static/gallery/"
const configPath = "config/"

var subSites []string

type Page struct {
	Title       string
	Description string
	OrigImgDir  string
	PrevImgDir  string
	ThumbImgDir string
	FeatImgDir  string
	GalleryPath string
	Images      map[string]bool
	Download    bool
}

func initPage() Page {
	return Page{
		OrigImgDir:  origImgDir,
		PrevImgDir:  prevImgDir,
		ThumbImgDir: thumbImgDir,
		FeatImgDir:  featImgDir,
		GalleryPath: galleryPath,
	}
}

func initSubSites() {
	files := readDir(galleryPath)

	for _, file := range files {
		if file.IsDir() {
			subSites = append(subSites, file.Name()+"/")
		}
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

func readYAML(filename string) (*Page, error) {
	page := initPage()
	source, err := ioutil.ReadFile(configPath + filename + ".yaml")
	if err != nil {
		panic(err)
	}

	err = yaml.Unmarshal(source, &page)
	if err != nil {
		panic(err)
	}

	page.Images = make(map[string]bool)

	for _, orig := range readDir(galleryPath + page.Title + "/" + origImgDir) {
		page.Images[orig.Name()] = false
	}
	for _, image := range readDir(galleryPath + page.Title + "/" + featImgDir) {
		page.Images[image.Name()] = true
	}

	return &page, err
}

func galleryHandler(w http.ResponseWriter, r *http.Request) {
	title := strings.Replace(r.URL.Path, "/", "", 2)

	if recreateZip {
		folders := []string{galleryPath + title + "/" + origImgDir,galleryPath + title + "/" +featImgDir}
		addZip(galleryPath+title+"/"+title+"_images.zip", folders)
	}

	p, _ := readYAML(title)
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

	for _, file := range readDir(configPath) {
		http.HandleFunc("/"+strings.Replace(file.Name(), ".yaml", "", 1)+"/", galleryHandler)
	}

	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func main() {
	initSubSites()
	checkSubSites(subSites)

	go watchFile(subSites)
	initWebServer("8080")
	//subSite := "ungarn/"
	//folders := []string{galleryPath + subSite + origImgDir,galleryPath + subSite +featImgDir}
	//addZip(galleryPath+subSite+strings.Replace(subSite,"/","",1)+"_images.zip", folders)
}
