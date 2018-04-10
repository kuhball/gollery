package gollery

import (
	"io/ioutil"
	"log"
	"sort"
	"time"

	"gopkg.in/yaml.v2"
)

const origImgDir = "img/"
const prevImgDir = "preview/"
const featImgDir = "featured/"
const thumbImgDir = "thumbnail/"

const thumbSize = 396
const featSize = 796
const prevSize = 1080

// GlobConfig contains the read config from the config.yaml
var GlobConfig Config
var galleryPath = getDir() + "/"

// Config Struct for the config.yaml with Port and all the galleries.
type Config struct {
	Port      string
	Galleries map[string]*Gallery
}

// Gallery Struct for a gallery within the config struct.
type Gallery struct {
	Title       string
	Description string
	Download    bool
	Images      []Image `yaml:"-"`
	Dir         dir     `yaml:"-"`
}

// Image Struct for image within gallery struct.
type Image struct {
	Name    string
	Date    string
	Time    time.Time
	Ratio   float32
	Feature bool
}

// Struct for all the paths - needed in the html template
type dir struct {
	OrigImgDir  string
	PrevImgDir  string
	ThumbImgDir string
	FeatImgDir  string
	GalleryPath string
}

// Initialize the paths for the images
func initDir() dir {
	return dir{
		OrigImgDir:  origImgDir,
		PrevImgDir:  prevImgDir,
		ThumbImgDir: thumbImgDir,
		FeatImgDir:  featImgDir,
		GalleryPath: "image/",
	}
}

// ReadConfig - Checks whether provided config path is valid
// Read & unmarshal config.yaml
// Call initImages for all galleries from the config file
// TODO: add support for json config file
func ReadConfig(f string, initialize bool) Config {
	if checkFile(f) {
		log.Fatal(f + ": Does not exist.")
	}

	var c Config
	source, err := ioutil.ReadFile(f)
	check(err)

	err = yaml.Unmarshal(source, &c)
	check(err)
	if initialize {
		for gallery := range c.Galleries {
			initImages(c, gallery)
		}
	}

	return c
}

// Read all origImgDir & featImgDir for the given gallery folder
// Call returnImageData() for every image
// Write image data into given config struct / gallery struct / image struct
func initImages(c Config, gallery string) Config {
	for _, orig := range readDir(gallery + "/" + origImgDir) {
		date, tm, ratio := returnImageData(gallery + "/" + origImgDir + orig.Name())
		c.Galleries[gallery].Images = append(c.Galleries[gallery].Images, Image{Name: orig.Name(), Date: date, Time: tm, Feature: false, Ratio: ratio})
	}
	for _, orig := range readDir(gallery + "/" + featImgDir) {
		date, tm, ratio := returnImageData(gallery + "/" + featImgDir + "/" + orig.Name())
		c.Galleries[gallery].Images = append(c.Galleries[gallery].Images, Image{Name: orig.Name(), Date: date, Time: tm, Feature: true, Ratio: ratio})
	}
	sort.SliceStable(c.Galleries[gallery].Images, func(i, j int) bool {
		return c.Galleries[gallery].Images[i].Time.Before(c.Galleries[gallery].Images[j].Time)
	})
	return c
}
