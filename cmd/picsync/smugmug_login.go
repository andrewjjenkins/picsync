package main

import (
	"fmt"
	"os"

	"github.com/andrewjjenkins/picsync/pkg/smugmug"
	"github.com/spf13/cobra"
)

var (
	smugmugLogin = &cobra.Command{
		Use:   "login",
		Short: "Complete one-time log in to SmugMug (OAuth)",
		Run:   runSmugmugLogin,
	}

	smugmugLoginOut string
)

func init() {
	smugmugLogin.PersistentFlags().StringVarP(
		&smugmugLoginOut,
		"outfile",
		"o",
		"",
		"Write token config out to file (like picsync-config.yaml)",
	)

	rootCmd.AddCommand(smugmugLogin)
}

func runSmugmugLogin(cmd *cobra.Command, args []string) {
	consumerKey := viper.GetString("smugmug_api_key")
	if consumerKey == "" {
		panic("Must provide a smugmug API key")
	}
	consumerSecret := viper.GetString("smugmug_api_secret")
	if consumerSecret == "" {
		panic("Must provide a smugmug API secret")
	}
	auth, err := smugmug.Login(consumerKey, consumerSecret)
	if err != nil {
		fmt.Printf("Login error: %v", err)
	}

	toWrite := fmt.Sprintf(
		"# Keep this file confidential.\n"+
			"# If you lose it, de-authorize nixplay-sync from your SmugMug account and repeat 'picsync login'\n"+
			"smugmug:\n"+
			"  access:\n"+
			"    token: \"%s\"\n"+
			"    secret: \"%s\"\n"+
			"  consumer:\n"+
			"    token: \"%s\"\n"+
			"    secret: \"%s\"\n",
		auth.Token,
		auth.Secret,
		consumerKey,
		consumerSecret,
	)

	var outfile *os.File
	if smugmugLoginOut == "-" || smugmugLoginOut == "" {
		outfile = os.Stdout
	} else {
		outfile, err = os.Create(smugmugLoginOut)
		if err != nil {
			panic(err)
		}
		defer outfile.Close()
	}
	_, err = outfile.WriteString(toWrite)
	if err != nil {
		fmt.Printf("Error writing login info to %s: %v\n", smugmugLoginOut, err)
	}
	outfile.Sync()
}
