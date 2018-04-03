package gollery

import (
	"net/http"
	"os"
	"log"
	"strings"
	"html/template"
	"path/filepath"
)

func galleryHandler(w http.ResponseWriter, r *http.Request) {
	title := strings.Replace(r.URL.Path, "/", "", 2)

	if recreate {
		folders := []string{filepath.FromSlash(galleryPath + title + "/" + origImgDir), filepath.FromSlash(galleryPath + title + "/" + featImgDir)}
		addZip(filepath.FromSlash(galleryPath + title + "/" + title + "_images.zip"), folders)
		getFeatured(GlobConfig, title)
	}

	GlobConfig.Galleries[title].Dir = initDir()

	t, _ := template.ParseFiles(filepath.FromSlash(webPath + "template/gallery.html"))
	t.Execute(w, GlobConfig.Galleries[title])
}

func staticHandler(w http.ResponseWriter, r *http.Request) {
	path := webPath + "/" + r.URL.Path
	if f, err := os.Stat(path); err == nil && !f.IsDir() {
		http.ServeFile(w, r, filepath.FromSlash(path))
		return
	}

	http.NotFound(w, r)
}

func imageHandler(w http.ResponseWriter, r *http.Request) {
	path := filepath.FromSlash(getDir() + r.URL.Path[len("/image"):])

	if f, err := os.Stat(path); err == nil && !f.IsDir() && !strings.Contains(path, "config") {
		http.ServeFile(w, r, path)
		return
	}

	http.NotFound(w, r)
}

func initWebServer(port string) {
	http.HandleFunc("/static/", staticHandler)
	http.HandleFunc("/image/", imageHandler)

	for subSite := range GlobConfig.Galleries {
		http.HandleFunc("/"+subSite, galleryHandler)
	}

	log.Fatal(http.ListenAndServe(":"+port, nil))
}
