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
	Images      map[string]Image `yaml:"-"`
	Dir         dir              `yaml:"-"`
}

type Image struct {
	Feature  bool
	Date     string
	Ratio	 float32
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
	c.Galleries[subSite].Images = make(map[string]Image)
	for _, orig := range readDir(subSite + "/" + origImgDir) {
		date, ratio := returnImageData(subSite + "/" + origImgDir  + orig.Name())
		c.Galleries[subSite].Images[orig.Name()] = Image{Date: date, Feature: false, Ratio: ratio}
	}
	for _, orig := range readDir(subSite + "/" + featImgDir) {
		date, ratio := returnImageData(subSite + "/" + featImgDir + "/" + orig.Name())
		c.Galleries[subSite].Images[orig.Name()] = Image{Date: date, Feature: true, Ratio: ratio}
	}
	//var previousDate string
	//for _, image := range c.Galleries[subSite].Images {
	//	if image.Date == previousDate {
	//		image.Date = ""
	//	} else {
	//		previousDate = image.Date
	//	}
	//}
	//for image := range c.Galleries[subSite].Images {
	//	log.Print(image)
	//}
	return c
}
