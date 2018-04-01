package main

import (
	"github.com/scouball/gollery"
)

func main() {
	//gollery.GlobConfig = gollery.ReadConfig("",true)
	//go initWebServer(globConfig.Port)
	//checkSubSites(globConfig.Galleries)
	//
	//watchFile(globConfig.Galleries)
	gollery.CliAccess()
	//initGollery("test")
}
