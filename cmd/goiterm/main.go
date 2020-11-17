package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name: "goiterm",
		Commands: []*cli.Command{
			{
				Name:        "install",
				Usage:       "goiterm install <go binary name>",
				Description: "Installs your plugin to the iTerm app",
				Action: func(c *cli.Context) error {
					bin := c.Args().First()
					if bin == "" {
						return cli.NewExitError("must pass go binary as first argument", 1)
					}
					homedir, err := os.UserHomeDir()
					if err != nil {
						return fmt.Errorf("os.UserHomeDir: %w", err)
					}
					p := filepath.Join(homedir, "/Library/Application Support/iTerm2/Scripts/", bin+".py")
					f := fmt.Sprintf(pyFile, bin)
					err = ioutil.WriteFile(p, []byte(f), 0666)
					if err != nil {
						return fmt.Errorf("ioutil.WriteFile: %w", err)
					}
					return nil
				},
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

const pyFile = `#!/usr/bin/env python3.7

import subprocess

subprocess.run(["%s"])
`
