package main

import (
	"fmt"
	"os"

	"github.com/andrewjjenkins/picsync/pkg/nixplay"
)

func getNixplayClientOrExit() (c nixplay.Client) {
	username := viper.GetString("nixplay.username")
	if username == "" {
		fmt.Printf("Must provide a nixplay username")
		os.Exit(1)
	}
	password := viper.GetString("nixplay.password")
	if password == "" {
		fmt.Printf("Must provide a nixplay password")
		os.Exit(1)
	}
	c, err := nixplay.NewClient(username, password, promReg)
	if err != nil {
		fmt.Printf("Nixplay login error: %v", err)
		os.Exit(1)
	}
	return c
}

func writeLoginOut(toWrite string) {
	var outfile *os.File
	var err error
	if loginOut == "-" || loginOut == "" {
		outfile = os.Stdout
	} else {
		outfile, err = os.Create(loginOut)
		if err != nil {
			panic(err)
		}
		defer outfile.Close()
	}
	_, err = outfile.WriteString(toWrite)
	if err != nil {
		fmt.Printf("Error writing login info to %s: %v\n", loginOut, err)
	}
	outfile.Sync()
}
