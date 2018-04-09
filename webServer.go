package gollery

import (
	"net/http"
	"os"
	"log"
	"strings"
	"path/filepath"
	"github.com/NYTimes/gziphandler"
	bTemplate "github.com/arschles/go-bindata-html-template"
)

var t *bTemplate.Template

type justFilesFilesystem struct {
	Fs http.FileSystem
}

// Read the html template from file into global variable and add a minus function.
// (only loaded once per start, not per request)
// TODO: integrate template into binary
//func initTemplate() {
//	var err error
//	t, err = template.New("gallery.html").Funcs(template.FuncMap{
//		"minus": func(a, b int) int { return a - b },
//	}).ParseFiles(webPath + "template/gallery.html")
//	check(err)
//}

func initTemplate() {
	var err error
	t, err = bTemplate.New("gallery.html", Asset).Funcs(bTemplate.FuncMap{
		"minus": func(a, b int) int { return a - b },
	}).Parse("web/template/gallery.html")
	if err != nil {
		log.Fatalf("error parsing template: %s", err)
	}
}

//
func galleryHandler() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		title := strings.Replace(r.URL.Path, "/", "", 2)

		if recreate {
			folders := []string{filepath.FromSlash(galleryPath + title + "/" + origImgDir), filepath.FromSlash(galleryPath + title + "/" + featImgDir)}
			addZip(filepath.FromSlash(galleryPath+title+"/"+title+"_images.zip"), folders)
			GlobConfig.Galleries[title].Images = GlobConfig.Galleries[title].Images[:0]
			initImages(GlobConfig, title)
			recreate = false
		}

		var err error
		err = t.Execute(w, GlobConfig.Galleries[title])
		check(err)
	})
}

// Handler for static files
// Only displays files, no folders
//func staticHandler(w http.ResponseWriter, r *http.Request) {
//	path := webPath + "/" + r.URL.Path
//	if f, err := os.Stat(path); err == nil && !f.IsDir() {
//		http.ServeFile(w, r, filepath.FromSlash(path))
//		return
//	}
//
//	http.NotFound(w, r)
//}

// Handler for all the image files within the gallery root folder
// Only displays files, no folders or config files
// TODO: return a real HTTP 404 error in case of not found.
func imageHandler(w http.ResponseWriter, r *http.Request) {
	path := filepath.FromSlash(getDir() + r.URL.Path[len("/image"):])

	if f, err := os.Stat(path); err == nil && !f.IsDir() && !strings.Contains(path, "config") {
		http.ServeFile(w, r, path)
		return
	}

	http.NotFound(w, r)
}

// Function for returning error for http folders and only serving files
func (fs justFilesFilesystem) Open(name string) (http.File, error) {
	f, err := fs.Fs.Open(name)

	if err != nil {
		return nil, err
	}

	stat, err := f.Stat()
	if stat.IsDir() {
		return nil, os.ErrNotExist
	}

	return f, nil
}

// Initializes the HTML template
// Registers static, image Handler
// Iterates over all galleries within the global config and registers a handler for each gallery
// Starts the http server on the configured port in the config.yaml
// TODO: register new gallery while service is running
func initWebServer(port string) {
	go initTemplate()

	fs := justFilesFilesystem{assetFS()}
	http.Handle("/static/", http.FileServer(fs))
	http.HandleFunc("/image/", imageHandler)

	for subSite := range GlobConfig.Galleries {
		GlobConfig.Galleries[subSite].Dir = initDir()
		http.Handle("/"+subSite, gziphandler.GzipHandler(galleryHandler()))
	}

	log.Fatal(http.ListenAndServe(":"+port, nil))
}
