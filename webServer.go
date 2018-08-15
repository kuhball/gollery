package gollery

import (
	"crypto/subtle"
	"github.com/NYTimes/gziphandler"
	bTemplate "github.com/arschles/go-bindata-html-template"
	"github.com/sethvargo/go-password/password"
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

// Generates Randomness for URL and Passwords
func getCrypto(len int) string {
	res, err := password.Generate(len, 10, 0, false, false)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf(res)
	return res
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
func galleryHandler(title, username, password, realm string) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if GlobConfig.Auth {
			user, pass, ok := r.BasicAuth()

			if !ok || subtle.ConstantTimeCompare([]byte(user), []byte(username)) != 1 || subtle.ConstantTimeCompare([]byte(pass), []byte(password)) != 1 {
				w.Header().Set("WWW-Authenticate", `Basic realm="`+realm+`"`)
				w.WriteHeader(401)
				w.Write([]byte("Unauthorised.\n"))
				return
			}
		}

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

func createGalleryHandle(subSite Gallery) {
	http.Handle("/"+subSite.Link, gziphandler.GzipHandler(galleryHandler(subSite.Title, "gollery", subSite.Password, "Please Login.")))
}

// Initializes the HTML template
// Registers static, image Handler
// Iterates over all galleries within the global config and registers a handler for each gallery
// Starts the http server on the configured port in the config.yaml
// TODO: HTTP2 PUSH - only available with TLS
func initWebServer(port string) {
	go initTemplate()

	fs := justFilesFilesystem{assetFS()}
	http.Handle("/", http.FileServer(fs))
	http.HandleFunc("/image/", imageHandler)

	for _, subSite := range GlobConfig.Galleries {
		subSite.Dir = initDir()
		createGalleryHandle(*subSite)
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
