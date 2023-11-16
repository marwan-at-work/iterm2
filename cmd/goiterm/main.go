package main

import (
	"fmt"
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
				Description: "Installs your plugin to iTerm2",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "name",
						Usage: "alternative plugin name instead of the binary",
					},
				},
				Action: func(c *cli.Context) error {
					bin := c.Args().First()
					if bin == "" {
						return cli.Exit("must pass go binary as first argument", 1)
					}
					appName := c.String("name")
					if appName == "" {
						appName = bin
					}
					homedir, err := os.UserHomeDir()
					if err != nil {
						return fmt.Errorf("os.UserHomeDir: %w", err)
					}
					p := filepath.Join(homedir, "/Library/Application Support/iTerm2/Scripts/", appName+".py")
					f := fmt.Sprintf(pyFile, bin)
					err = os.WriteFile(p, []byte(f), 0666)
					if err != nil {
						return fmt.Errorf("os.WriteFile: %w", err)
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
