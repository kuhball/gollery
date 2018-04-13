//Package gollery - main package of the application with all the logic
package gollery

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/manifoldco/promptui"
	"github.com/urfave/cli"
	"gopkg.in/yaml.v2"
)

// Function for creating a new root gallery folder
// Checks whether a custom path was specified and uses current path if not.
// Creates a custom_css folder, the config.yaml with a example gallery and the corresponding file structure
// TODO: import the custom_css files into html template
func initGollery(path string) error {
	//Define where the new gollery should be initialized
	if path == "" {
		pathSelect := promptui.Select{
			Label: "You haven't specified a Path. Should the new Gollery be initialized at " + getDir() + "?",
			Items: []string{"yep, go!", "nop!"},
		}
		enterPath := promptui.Prompt{
			Label: "Please enter a custom Path (full or starting at the current location)",
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
		err := os.Mkdir(path, 0600)
		check(err)
		log.Print("Created new Directory " + path + ".")
	}

	emptyPath := promptui.Select{
		Label: "The provided folder is not empty. Do you want to continue?",
		Items: []string{"yep, go!", "nop!"},
	}

	if len(readDir(path)) != 0 {
		if _, s, err := emptyPath.Run(); err != nil {
			fmt.Printf("Prompt failed %v\n", err)
			return err
		} else if s == "nop!" {
			return nil
		}
	}

	//if checkFile("custom_css") {
	//	err := os.Mkdir(path+"/custom_css", 0600)
	//	check(err)
	//	log.Print("Created folder custom_css.")
	//} else {
	//	log.Println("costum_css folder is already existing.")
	//}

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
			log.Print("Successfully overwritten config.yaml.")
		} else if s == "nop!" {
			log.Print("Old config.yaml wasn't changed.")
			return nil
		}
	} else {
		writeConfig(path, initExampleConfig())
		log.Print("Created new config.yaml.")
	}

	createGalleries(path)

	genImages := promptui.Select{
		Label: "Do you want some example Images?",
		Items: []string{"yep, go!", "nop!"},
	}

	if _, s, err := genImages.Run(); err != nil {
		fmt.Printf("Prompt failed %v\n", err)
		return err
	} else if s == "yep, go!" {
		downloadFile(galleryPath+"example/"+origImgDir+"example1.jpg", "https://unsplash.com/photos/H4Sv_zRXBos/download?force=true")
		downloadFile(galleryPath+"example/"+origImgDir+"example2.jpg", "https://unsplash.com/photos/bF9kRBJhMpE/download?force=true")
		downloadFile(galleryPath+"example/"+origImgDir+"example3.jpg", "https://unsplash.com/photos/XqMjjuQuyZQ/download?force=true")
	}

	log.Print("New gollery was created successfully 👍🏻")

	return nil
}

// Creates an example gallery configuration
func initExampleConfig() Config {
	g := make(map[string]*Gallery)
	e := Gallery{Title: "example", Description: "This is an example gallery.", Download: false, CustomCss: false}
	g["example"] = &e
	c := Config{Port: "8080", Galleries: g}

	return c
}

// Write a new config to the filesystem.
func writeConfig(path string, c Config) {
	d, err := yaml.Marshal(&c)
	check(err)

	err = ioutil.WriteFile(path+"/config.yaml", d, 0644)
	check(err)
}

// Function creates a new gallery within an existing root folder and config.yaml
// Reads the existing config file, asks for Title (unique), Description and Download (bool)
// Writes new config and generates folder structure for new gallery
func newGallery(path string) error {
	var err error
	var newData Gallery

	c := ReadConfig(path+"/config.yaml", false)

	title := promptui.Prompt{
		Label: "Title",
	}

	description := promptui.Prompt{
		Label: "Description",
	}

	download := promptui.Select{
		Label: "Compress all images automatically and provide a download button?",
		Items: []string{"yep, go!", "nop!"},
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

	if _, s, err := download.Run(); err != nil {
		fmt.Printf("Prompt failed %v\n", err)
		return err
	} else if s == "yep, go!" {
		newData.Download = true
	} else if s == "nop!" {
		newData.Download = false
	}

	c.Galleries[newData.Title] = &newData

	writeConfig(path, c)
	createGalleries(path)

	return nil
}

// Reads config.yaml from filesystem
// Checks whether new gallery name already has a folder -> aborts if yes
// Create all necessary subfolders for the gallery
func createGalleries(path string) {
	c := ReadConfig(path+"/config.yaml", false)

	for subsite := range c.Galleries {
		if checkFile(path + "/" + subsite) {
			err := os.Mkdir(path+"/"+subsite, 0755)
			check(err)
			log.Print("Created new folder " + subsite + ".")
		}
		if checkFile(path + "/" + subsite + "/img") {
			err := os.Mkdir(path+"/"+subsite+"/img", 0755)
			check(err)
			log.Print("Created new folder " + subsite + "/img .")
		}
		if checkFile(path + "/" + subsite + "/featured") {
			err := os.Mkdir(path+"/"+subsite+"/featured", 0755)
			check(err)
			log.Print("Created new folder " + subsite + "/featured .")
		}
		if checkFile(path + "/" + subsite + "/preview") {
			err := os.Mkdir(path+"/"+subsite+"/preview", 0755)
			check(err)
			log.Print("Created new folder " + subsite + "/preview .")
		}
		if checkFile(path + "/" + subsite + "/thumbnail") {
			err := os.Mkdir(path+"/"+subsite+"/thumbnail", 0755)
			check(err)
			log.Print("Created new folder " + subsite + "/thumbnail .")
		}
		createCustomCss(c, subsite)
	}
}

func startGollery(c *cli.Context, directory string) error {
	if c.Bool("webserver") && c.Bool("filewatcher") {
		log.Fatal("flag combination is not allowed")
	}

	if directory == "" {
		configPath = getDir() + "/config.yaml"
	} else {
		configPath = directory + "/config.yaml"
	}

	GlobConfig = ReadConfig(configPath, true)

	if c.Bool("webserver") && !c.Bool("filewatcher") {
		initWebServer(GlobConfig.Port)
	}
	if !c.Bool("webserver") && c.Bool("filewatcher") {
		go checkImageTool()
		checkSubSites(GlobConfig.Galleries)
		watchFile(GlobConfig.Galleries)
	}
	if !c.Bool("webserver") && !c.Bool("filewatcher") {
		go initWebServer(GlobConfig.Port)
		checkImageTool()
		checkSubSites(GlobConfig.Galleries)
		watchFile(GlobConfig.Galleries)
	}
	return nil
}

// CliAccess - Main function for all functionality
// provides all cli arguments via cli plugin - read doc for more information
//TODO: Remove gallery from config file
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
				return startGollery(c, directory)
			},
			Flags: []cli.Flag{
				cli.StringFlag{Name: "directory, d", Usage: "root path for gollery", Destination: &directory},
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
					return newGallery(getDir())
				}
				return newGallery(customDir)
			},
		},
	}

	err := app.Run(os.Args)
	check(err)
}
