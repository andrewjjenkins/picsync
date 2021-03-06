package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/andrewjjenkins/picsync/pkg/nixplay"
	"github.com/andrewjjenkins/picsync/pkg/smugmug"
)

func getSmugmugClientOrExit() (c *http.Client) {
	auth := smugmug.AccessAuth{
		Token:          viper.GetString("smugmug.access.token"),
		Secret:         viper.GetString("smugmug.access.secret"),
		ConsumerKey:    viper.GetString("smugmug.consumer.token"),
		ConsumerSecret: viper.GetString("smugmug.consumer.secret"),
	}
	if auth.Token == "" || auth.Secret == "" ||
		auth.ConsumerKey == "" || auth.ConsumerSecret == "" {
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
