package gollery

import (
	"net/http"
	"os"
	"log"
	"strings"
	"html/template"
)

func galleryHandler(w http.ResponseWriter, r *http.Request) {
	title := strings.Replace(r.URL.Path, "/", "", 2)

	if recreate {
		folders := []string{galleryPath + title + "/" + origImgDir, galleryPath + title + "/" + featImgDir}
		addZip(galleryPath+title+"/"+title+"_images.zip", folders)
		getFeatured(GlobConfig, title)
	}

	GlobConfig.Galleries[title].Dir = initDir()

	t, _ := template.ParseFiles(appPath + "web/template/gallery.html")
	t.Execute(w, GlobConfig.Galleries[title])
}

func staticHandler(w http.ResponseWriter, r *http.Request) {
	path := appPath + "/web/" + r.URL.Path
	if f, err := os.Stat(path); err == nil && !f.IsDir() {
		http.ServeFile(w, r, path)
		return
	}

	http.NotFound(w, r)
}

func imageHandler(w http.ResponseWriter, r *http.Request) {
	path := getDir() + r.URL.Path[len("/image"):]

	log.Println(path)
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
