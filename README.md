# gollery
A simple web gallery written in [golang](https://golang.org/) for serving images

## Features

- create different image sizes (thumbnail / preview / *future: mobile*) using [imagemagick](http://www.imagemagick.org/script/index.php)
- watch image folders and create images automatically
- serve a site for every gallery 
- responsive masonry layout
- create zip with all images
- custom css for every gallery

![alt text](screenshots/example_gollery.png "example gollery")

## Installation

### Prerequisites

For creating thumbnails & previews gollery uses imagemagick. Please install a suitable [imagemagick](http://www.imagemagick.org/script/download.php) version for your os and make sure it's reachable via `convert` or `magick`.

[Download imagemagick](http://www.imagemagick.org/script/download.php)

### Build from source

1. clone the github repo
2. `make install`
3. `make build`
4. there should be a application in your $GOPATH/bin called **gollery**

### Download binaries

// Provide binaries for linux & windows in the future

## Usage

### CLI

Gollery comes with a simple cli and 3 basic commands:

1. `gollery start`

   This command starts the webserver and the filewatcher. 

2. `gollery init`

   This command creates a new root folder with a `config.yaml` and a example gallery

3. `gollery new`

   This command creates a new gallery within an existing config.yaml and adds it to the config.
4. `gollery remove`

   This command removes a gallery from an existing config.yaml and deletes the folder structure.

The folder of the config.yaml and the galleries can be provided manual with `gollery -c /PATH/TO/GOLLERY/ COMMAND`. This is possible with all commands.
Check the help for further options.

### Docker

There's the possibility to build gollery within a docker container. *Needs still some work.*
