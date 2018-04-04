package gollery

import (
	"net/http"
	"os"
	"log"
	"strings"
	"html/template"
	"path/filepath"
	"github.com/NYTimes/gziphandler"
)

var t *template.Template

func initTemplate() {

	var err error
	t, err = template.New("gallery.html").Funcs(template.FuncMap{
		"minus": func(a, b int) int { return a - b },
	}).ParseFiles(webPath + "template/gallery.html")
	check(err)
}

func galleryHandler() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		title := strings.Replace(r.URL.Path, "/", "", 2)

		if recreate {
			folders := []string{filepath.FromSlash(galleryPath + title + "/" + origImgDir), filepath.FromSlash(galleryPath + title + "/" + featImgDir)}
			addZip(filepath.FromSlash(galleryPath+title+"/"+title+"_images.zip"), folders)
			GlobConfig.Galleries[title].Images = GlobConfig.Galleries[title].Images[:0]
			getFeatured(GlobConfig, title)
		}

		var err error
		err = t.Execute(w, GlobConfig.Galleries[title])
		check(err)
	})
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
	go initTemplate()

	http.HandleFunc("/static/", staticHandler)
	http.HandleFunc("/image/", imageHandler)

	for subSite := range GlobConfig.Galleries {
		GlobConfig.Galleries[subSite].Dir = initDir()
		http.Handle("/"+subSite, gziphandler.GzipHandler(galleryHandler()))
	}

	log.Fatal(http.ListenAndServe(":"+port, nil))
}
