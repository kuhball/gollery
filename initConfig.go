package gollery

import (
	"io/ioutil"
	"gopkg.in/yaml.v2"
	"log"
	"sort"
	"time"
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
	Images      []Image `yaml:"-"`
	Dir         dir     `yaml:"-"`
}

type Image struct {
	Name    string
	Feature bool
	Date    string
	Time    time.Time
	Ratio   float32
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
	check(err)

	err = yaml.Unmarshal(source, &c)
	check(err)
	if feature {
		for subSite := range c.Galleries {
			getFeatured(c, subSite)
		}
	}

	return c
}

func getFeatured(c Config, subSite string) Config {
	for _, orig := range readDir(subSite + "/" + origImgDir) {
		date, tm, ratio := returnImageData(subSite + "/" + origImgDir + orig.Name())
		c.Galleries[subSite].Images = append(c.Galleries[subSite].Images, Image{Name: orig.Name(), Date: date, Time: tm, Feature: false, Ratio: ratio})
	}
	for _, orig := range readDir(subSite + "/" + featImgDir) {
		date, tm, ratio := returnImageData(subSite + "/" + featImgDir + "/" + orig.Name())
		c.Galleries[subSite].Images = append(c.Galleries[subSite].Images, Image{Name: orig.Name(), Date: date, Time: tm, Feature: true, Ratio: ratio})
	}
	sort.SliceStable(c.Galleries[subSite].Images, func(i, j int) bool { return c.Galleries[subSite].Images[i].Time.Before(c.Galleries[subSite].Images[j].Time) })
	return c
}
