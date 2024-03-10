package main

import (
	"flag"
	"os"
	"os/user"
	"path/filepath"
)

var (
	configFile string
)

func init() {

	usr, err := user.Current()

	if err != nil {
		os.Stderr.WriteString(err.Error())
		os.Exit(1)
	}

	flag.StringVar(&configFile, "config", filepath.Join(usr.HomeDir, "/.config/icpm-control"), "Application config location")
}
