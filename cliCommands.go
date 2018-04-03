package gollery

import (
	"os"
	"github.com/urfave/cli"
	"log"
	"errors"
	"fmt"
	"github.com/manifoldco/promptui"
	"strconv"
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

func initGollery(path string) error {
	if path == "" {
		pathSelect := promptui.Select{
			Label: "You haven't specified a Path. Should the new Gollery be initialized at " + getDir() + "?",
			Items: []string{"yep, go!", "nop!"},
		}
		enterPath := promptui.Prompt{
			Label: "Please enter a custom Path (full)",
		}

		if _, s, err := pathSelect.Run(); err != nil {
			fmt.Printf("Prompt failed %v\n", err)
			return err
		} else if s == "yep, go!" {
			path = getDir()
		} else if s == "nop!" {
		checkPath:
			var err error
			if path, err = enterPath.Run(); err != nil {
				fmt.Printf("Prompt failed %v\n", err)
				return err
			}

			if checkFile(path) {
				log.Println("The provided path doesn't exist. Please try again.")
				goto checkPath
			}
		}
	}

	if checkFile(path) {
		os.Mkdir(path, 600)
	} else {
		log.Println("Directory is already existing.")
	}

	if checkFile("custom_css") {
		os.Mkdir(path+"/custom_css", 600)
	} else {
		log.Println("costum_css folder is already existing.")
	}

	if !checkFile(path + "/config.yaml") {
		overwriteConfig := promptui.Select{
			Label: "The config.yaml file already exists. Do you want to overwrite it?",
			Items: []string{"yep, go!", "nop!"},
		}

		if _, s, err := overwriteConfig.Run(); err != nil {
			fmt.Printf("Prompt failed %v\n", err)
			return err
		} else if s == "yep, go!" {
			writeConfig(path, initExampleConfig())
		} else if s == "nop!" {
			return nil
		}
	} else {
		writeConfig(path, initExampleConfig())
	}

	createGalleries(path)

	return nil
}

func initExampleConfig() Config {
	g := make(map[string]*Galleries)
	e := Galleries{Title: "example", Description: "This is an example gallery.", Download: false}
	g["example"] = &e
	c := Config{Port: "8080", Galleries: g}

	return c
}

func writeConfig(path string, c Config) {
	d, err := yaml.Marshal(&c)
	check(err)

	err = ioutil.WriteFile(path+"/config.yaml", d, 0644)
	check(err)
}

func newPrompts(path string) error {
	var err error
	var s string
	var newData Galleries

	c := ReadConfig(path+"/config.yaml", false)

	title := promptui.Prompt{
		Label: "Title",
	}

	description := promptui.Prompt{
		Label: "Description",
	}

	download := promptui.Select{
		Label: "Compress all images automatically and provide a download button?",
		Items: []bool{true, false},
	}

newTitle:
	if newData.Title, err = title.Run(); err != nil {
		fmt.Printf("Prompt failed %v\n", err)
		return err
	}

	if len(newData.Title) < 1 {
		log.Println("Title must have atleast 1 character, try again.")
		goto newTitle
	}

	if c.Galleries[newData.Title] != nil {
		log.Println("Gallery is already existing, try again.")
		goto newTitle
	}

	if newData.Description, err = description.Run(); err != nil {
		fmt.Printf("Prompt failed %v\n", err)
		return err
	}

	if _, s, err = download.Run(); err != nil {
		fmt.Printf("Prompt failed %v\n", err)
		return err
	}
	if newData.Download, err = strconv.ParseBool(s); err != nil {
		fmt.Printf("Boolean convertation failed %v\n", err)
	}

	c.Galleries[newData.Title] = &newData

	writeConfig(path, c)
	createGalleries(path)

	return nil
}

func createGalleries(path string) {
	c := ReadConfig(path+`\config.yaml`, false)

	log.Println(c)

	for subsite := range c.Galleries {
		if checkFile(path + "/" + subsite) {
			os.Mkdir(path+"/"+subsite, 600)
			os.Mkdir(path+"/"+subsite+"/img", 600)
			os.Mkdir(path+"/"+subsite+"/featured", 600)
			os.Mkdir(path+"/"+subsite+"/preview", 600)
			os.Mkdir(path+"/"+subsite+"/thumbnail", 600)
		} else {
			log.Println(subsite + " structure is already existing.")
		}
	}
}

func CliAccess() {
	var directory string
	var customDir string

	app := cli.NewApp()
	app.Name = "gollery"
	app.Version = "0.1.0"
	app.Usage = "start, initialize and create new galleries in gollery"
	app.Authors = []cli.Author{
		{
			Name: "Simon Couball", Email: "info@simoncouball.de",
		},
	}

	app.Flags = []cli.Flag{
		cli.StringFlag{Name: "custom-dir, c", Usage: "custom directory ", Destination: &customDir},
	}

	app.Commands = []cli.Command{
		{
			Name:        "start",
			Aliases:     []string{"s"},
			Usage:       "Start gollery as a daemon",
			Description: "moin",
			Action: func(c *cli.Context) error {
				if c.Bool("webserver") && c.Bool("filewatcher") {
					return errors.New("flag combination is not allowed")
				}

				if directory == "" {
					GlobConfig = ReadConfig(getDir()+"/config.yaml", true)
				} else {
					GlobConfig = ReadConfig(directory+"/config.yaml", true)
				}

				if c.String("webpath") == "" {
					webPath = getGoPath() + "/src/github.com/scouball/gollery/"
				} else {
					webPath = c.String("webpath")
				}

				log.Print(c.String("webpath"))

				initWebServer(GlobConfig.Port)
				//checkSubSites(GlobConfig.Galleries)

				//watchFile(GlobConfig.Galleries)
				return nil
			},
			Flags: []cli.Flag{
				cli.StringFlag{Name: "directory, d", Usage: "root path for gollery", Destination: &directory},
				cli.StringFlag{Name: "webpath, p", Usage: "custom location for web folder (needed for docker)"},
				cli.BoolFlag{Name: "webserver, w", Usage: "only start webserver"},
				cli.BoolFlag{Name: "filewatcher, f", Usage: "only start filewatcher"},
			},
		},
		{
			Name:        "init",
			Aliases:     []string{"i"},
			Usage:       "init new root directory",
			Description: "test",
			Action: func(c *cli.Context) error {
				return initGollery(customDir)
			},
		},
		{
			Name:        "new",
			Aliases:     []string{"n"},
			Usage:       "create new gallery",
			Description: "test",
			Action: func(c *cli.Context) error {
				if customDir == "" {
					return newPrompts(getDir())
				} else {
					return newPrompts(customDir)
				}
			},
		},
	}

	err := app.Run(os.Args)
	check(err)
}
