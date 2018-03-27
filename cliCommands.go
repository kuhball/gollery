package main

import (
	"os"
	"log"
	"github.com/urfave/cli"
	"fmt"
)

func cliAccess () {
	app := cli.NewApp()

	app.Action = func(c *cli.Context) error {
		fmt.Printf("Hello %q", c.Args().Get(0))
		return nil
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
