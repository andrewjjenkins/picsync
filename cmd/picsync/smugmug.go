package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/andrewjjenkins/picsync/pkg/smugmug"
	"github.com/spf13/cobra"
)

var (
	smugmugCmd = &cobra.Command{
		Use:   "smugmug",
		Short: "sync pictures from SmugMux to Nixplay",
		Run:   runSmugmug,
	}

	smugmugLogin = &cobra.Command{
		Use:   "login",
		Short: "Complete one-time log in to SmugMug (OAuth)",
		Run:   runSmugmugLogin,
	}

	smugmugSync = &cobra.Command{
		Use:   "sync <smugmug-album> <nixplay-album>",
		Short: "Sync a smugmug album to a nixplay album",
		Run:   runSmugmugSync,
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) != 2 || args[0] == "" || args[1] == "" {
				return errors.New("specify the name of the source SmugMug album and destination NixPlay album")
			}
			return nil
		},
	}

	syncEvery string
	maxPics   int
)

func init() {
	smugmugLogin.PersistentFlags().StringVarP(
		&loginOut,
		"outfile",
		"o",
		"",
		"Write token config out to file (like picsync-config.yaml)",
	)
	smugmugCmd.AddCommand(smugmugLogin)

	smugmugSync.PersistentFlags().StringVarP(
		&syncEvery,
		"every",
		"d",
		"",
		"Sync every interval (like \"30s\" or \"1h\")",
	)
	smugmugSync.PersistentFlags().IntVarP(
		&maxPics,
		"max",
		"n",
		100,
		"Maximum pictures to sync",
	)
	smugmugCmd.AddCommand(smugmugSync)

	rootCmd.AddCommand(smugmugCmd)
}

func runSmugmug(cmd *cobra.Command, args []string) {
	fmt.Println("Choose a Smugmug-related operation like \"picsync smugmug login\" or \"picsync smugmug sync\"")
	fmt.Println("(try --help to see options)")
	os.Exit(1)
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
	consumerAuth := &smugmug.ConsumerAuth{
		Token:  consumerKey,
		Secret: consumerSecret,
	}
	accessAuth, err := smugmug.Login(consumerKey, consumerSecret)
	if err != nil {
		fmt.Printf("Login error: %v", err)
	}
	auth := &smugmug.SmugmugAuth{
		Access:   *accessAuth,
		Consumer: *consumerAuth,
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
		auth.Access.Token,
		auth.Access.Secret,
		auth.Consumer.Token,
		auth.Consumer.Secret,
	)
	writeLoginOut(toWrite)

}

func runSmugmugSync(cmd *cobra.Command, args []string) {
	smugmugAlbumName := args[0]
	nixplayAlbumName := args[1]

	if syncEvery != "" {
		runSyncEvery(smugmugAlbumName, nixplayAlbumName, syncEvery)
	} else {
		runSyncOnce(smugmugAlbumName, nixplayAlbumName)
	}
}
