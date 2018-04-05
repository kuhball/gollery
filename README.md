# gollery
A simple web gallery written in [golang](https://golang.org/) for serving images 

## Features

- create different image sizes (thumbnail / preview / *future: mobile*)
- watch image folders and create images automatically
- serve a site for every gallery 
- responsive masonry layout

// here could be a demo image

## Installation

1. clone the github repo
2. `make build`
3. there shout be a application in your go path called **gollery**

// Provide binarys for linux & windows in the future

## Usage

### CLI

Gollery comes with a simple cli and 3 basic commands:

1. `gollery start`

   This command starts the webserver and the filewatcher. 

2. `gollery init`

   This command creates a new root folder with a `config.yaml` and a example gallery

3. `gollery new`

   This command creates a new gallery within an existing root folder and adds it to the config.

### Docker

There's the possibility to build gollery within a docker container. *Needs still some work.*
