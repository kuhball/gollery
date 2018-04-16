package gollery

import (
	"github.com/NYTimes/gziphandler"
	bTemplate "github.com/arschles/go-bindata-html-template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

var t *bTemplate.Template

type justFilesFilesystem struct {
	Fs http.FileSystem
}

// Read the html template from file into global variable and add a minus function.
// (only loaded once per start, not per request)
func initTemplate() {
	var err error
	t, err = bTemplate.New("gallery.html", Asset).Funcs(bTemplate.FuncMap{
		"minus": func(a, b int) int { return a - b },
	}).Parse("web/template/gallery.html")
	if err != nil {
		log.Fatalf("error parsing template: %s", err)
	}
}

//TODO: debug on uberspace - not sure what the issue is
func galleryHandler() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		title := strings.Replace(r.URL.Path, "/", "", 2)

		if recreate {
			addZip(GlobConfig, title)
			GlobConfig.Galleries[title].Images = GlobConfig.Galleries[title].Images[:0]
			initImages(GlobConfig, title)
			recreate = false
		}

		//if pusher, ok := w.(http.Pusher); ok {
		//	if err := pusher.Push("/static/css/custom.css", nil); err != nil {
		//		log.Printf("Failed to push: %v", err)
		//	}
		//	if err := pusher.Push("/static/js/index.js", nil); err != nil {
		//		log.Printf("Failed to push: %v", err)
		//	}
		//}

		var err error
		err = t.Execute(w, GlobConfig.Galleries[title])
		check(err)

	})
}

// Handler for all the image files within the gallery root folder
// Only displays files, no folders or config files
func imageHandler(w http.ResponseWriter, r *http.Request) {
	path := filepath.FromSlash(galleryPath + r.URL.Path[len("/image"):])

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
	check(err)
	if stat.IsDir() {
		return nil, os.ErrNotExist
	}

	return f, nil
}

func createGalleryHandle(c Config, subSite string) {
	c.Galleries[subSite].Dir = initDir()
	http.Handle("/"+subSite, gziphandler.GzipHandler(galleryHandler()))
}

// Initializes the HTML template
// Registers static, image Handler
// Iterates over all galleries within the global config and registers a handler for each gallery
// Starts the http server on the configured port in the config.yaml
// TODO: HTTP2 PUSH - only available with TLS
func initWebServer(port string) {
	go initTemplate()

	fs := justFilesFilesystem{assetFS()}
	http.Handle("/static/", http.FileServer(fs))
	http.HandleFunc("/image/", imageHandler)

	for subSite := range GlobConfig.Galleries {
		createGalleryHandle(GlobConfig, subSite)
	}
	log.Print("Starting webserver on Port " + port)
	if GlobConfig.SSL {
		if checkFile(GlobConfig.Cert) {
			log.Fatal("Cert file is not existing.")
		}
		if checkFile(GlobConfig.Key) {
			log.Fatal("Key file is not existing.")
		}
		log.Print("Starting TLS Webserver.")
		log.Fatal(http.ListenAndServeTLS(":"+port, GlobConfig.Cert, GlobConfig.Key, nil))
	} else {
		log.Fatal(http.ListenAndServe(":"+port, nil))
	}
}
