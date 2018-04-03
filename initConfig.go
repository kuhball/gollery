package gollery

import (
	"io/ioutil"
	"gopkg.in/yaml.v2"
	"log"
)

const origImgDir = "img/"
const prevImgDir = "preview/"
const featImgDir = "featured/"
const thumbImgDir = "thumbnail/"

const thumbSize = 396
const featSize = 796
const prevSize = 1080

var GlobConfig Config
var webPath string
var galleryPath = getDir() + "/"


type Config struct {
	Port      string
	Galleries map[string]*Galleries
}

type Galleries struct {
	Title       string
	Description string
	Download    bool
	Images      map[string]bool `yaml:"-"`
	Dir         dir				`yaml:"-"`
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
		GalleryPath: "image/",
	}
}

func ReadConfig(f string, feature bool) Config {
	if checkFile(f) {
		log.Fatal(f + ": Does not exist.")
	}

	var c Config
	source, err := ioutil.ReadFile(f)
	if err != nil {
		panic(err)
	}

	err = yaml.Unmarshal(source, &c)
	if err != nil {
		panic(err)
	}
	if feature {
		for subSite := range c.Galleries {
			getFeatured(c, subSite)
		}
	}

	return c
}

func getFeatured(c Config, subSite string) Config {
	c.Galleries[subSite].Images = make(map[string]bool)
	for _, orig := range readDir(subSite + "/" + origImgDir) {
		c.Galleries[subSite].Images[orig.Name()] = false
	}
	for _, orig := range readDir(subSite + "/" + featImgDir) {
		c.Galleries[subSite].Images[orig.Name()] = true
	}
	return c
}
