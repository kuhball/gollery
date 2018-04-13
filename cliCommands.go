//Package gollery - main package of the application with all the logic
package gollery

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"errors"
	"github.com/manifoldco/promptui"
	"github.com/urfave/cli"
	"gopkg.in/yaml.v2"
)

// Function for creating a new root gallery folder
// Checks whether a custom path was specified and uses current path if not.
// Creates a custom_css folder, the config.yaml with a example gallery and the corresponding file structure
func initGollery(path string) error {
	//Define where the new gollery should be initialized
	if path == "" {
		pathSelect := promptui.Select{
			Label: "You haven't specified a Path. Should the new Gollery be initialized at " + getDir() + "?",
			Items: []string{"yep, go!", "nop!"},
		}
		pathValidate := func(input string) error {
			if checkFile(input) {
				return errors.New("provided path doesn't exist")
			}
			return nil
		}
		enterPath := promptui.Prompt{
			Label:    "Please enter a custom Path (full or starting at the current location)",
			Validate: pathValidate,
		}

		if _, s, err := pathSelect.Run(); err != nil {
			fmt.Printf("Prompt failed %v\n", err)
			return err
		} else if s == "yep, go!" {
			path = getDir()
		} else if s == "nop!" {
			var err error
			if path, err = enterPath.Run(); err != nil {
				fmt.Printf("Prompt failed %v\n", err)
				return err
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

	titleValidate := func(input string) error {
		if len(input) < 1 {
			return errors.New("title must have at least 1 character")
		}
		if c.Galleries[input] != nil {
			return errors.New("gallery is already existing")
		}
		return nil
	}

	title := promptui.Prompt{
		Label:    "Title",
		Validate: titleValidate,
	}

	description := promptui.Prompt{
		Label: "Description",
	}

	download := promptui.Select{
		Label: "Compress all images automatically and provide a download button?",
		Items: []string{"yep, go!", "nop!"},
	}

	customCss := promptui.Select{
		Label: "Create a custom_css file for this gallery?",
		Items: []string{"yep, go!", "nop!"},
	}

	if newData.Title, err = title.Run(); err != nil {
		fmt.Printf("Prompt failed %v\n", err)
		return err
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

	if _, s, err := customCss.Run(); err != nil {
		fmt.Printf("Prompt failed %v\n", err)
		return err
	} else if s == "yep, go!" {
		newData.CustomCss = true
	} else if s == "nop!" {
		newData.CustomCss = false
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

func removeGallery(path string) error {
	c := ReadConfig(path+"/config.yaml", false)

	log.Print("Please provide the title of the gallery:")

	validate := func(input string) error {
		if c.Galleries[input] == nil {
			return errors.New("gallery is not existing")
		}
		return nil
	}

	title := promptui.Prompt{
		Label:    "Title",
		Validate: validate,
	}

	result, err := title.Run()
	check(err)

	if !checkFile(path + "/" + result) {
		removeFile(path + "/" + result)
	}

	delete(c.Galleries, result)
	writeConfig(path, c)

	return nil
}

// CliAccess - Main function for all functionality
// provides all cli arguments via cli plugin - read doc for more information
//TODO: test custom directory option
func CliAccess() {
	var directory string
	var customDir string

	app := cli.NewApp()
	app.Name = "gollery"
	app.Version = "0.1.0"
	app.Usage = "start, initialize, create and remove new galleries in gollery"
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
		{
			Name:        "remove",
			Aliases:     []string{"r"},
			Usage:       "remove a gallery",
			Description: "test",
			Action: func(c *cli.Context) error {
				if customDir == "" {
					return removeGallery(getDir())
				}
				return removeGallery(customDir)
			},
		},
	}

	err := app.Run(os.Args)
	check(err)
}
