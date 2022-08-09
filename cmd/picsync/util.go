package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/andrewjjenkins/picsync/pkg/nixplay"
	"github.com/andrewjjenkins/picsync/pkg/smugmug"
)

func getSmugmugClientOrExit() (c *http.Client) {
	auth := smugmug.SmugmugAuth{
		Access: smugmug.AccessAuth{
			Token:  viper.GetString("smugmug.access.token"),
			Secret: viper.GetString("smugmug.access.secret"),
		},
		Consumer: smugmug.ConsumerAuth{
			Token:  viper.GetString("smugmug.consumer.token"),
			Secret: viper.GetString("smugmug.consumer.secret"),
		},
	}
	if auth.Access.Token == "" || auth.Access.Secret == "" ||
		auth.Consumer.Token == "" || auth.Consumer.Secret == "" {
		fmt.Printf("No smugmug auth; do \"picsync login -o picsync-config.yaml\"\n")
		os.Exit(1)
	}
	client, err := smugmug.Access(&auth)
	if err != nil {
		fmt.Printf("Smugmug auth failed (%v); "+
			"repeat \"picsync login -o picsync-config.yaml\"\n", err)
		os.Exit(1)
	}
	return client
}

func getNixplayClientOrExit() (c *http.Client) {
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
	c, err := nixplay.Login(username, password)
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
