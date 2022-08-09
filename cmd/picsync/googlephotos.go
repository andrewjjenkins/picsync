package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/andrewjjenkins/picsync/pkg/googlephotos"
	"github.com/spf13/cobra"
	"golang.org/x/oauth2"
)

var (
	googlephotosCmd = &cobra.Command{
		Use:   "googlephotos",
		Short: "sync pictures from Google Photos to Nixplay",
		Run:   runGooglephotos,
	}

	googlephotosLogin = &cobra.Command{
		Use:   "login",
		Short: "Complete one-time log in to Google Photos (OAuth)",
		Run:   runGooglephotosLogin,
	}

	googlephotosList = &cobra.Command{
		Use:   "list [<albumId>]",
		Short: "List All albums, or the photos in a particular album (by id)",
		Run:   runGooglephotosList,
	}
)

func init() {
	googlephotosLogin.PersistentFlags().StringVarP(
		&loginOut,
		"outfile",
		"o",
		"",
		"Write token config out to file (like picsync-config.yaml)",
	)

	googlephotosCmd.AddCommand(googlephotosLogin)

	googlephotosCmd.AddCommand(googlephotosList)

	rootCmd.AddCommand(googlephotosCmd)
}

func runGooglephotos(cmd *cobra.Command, args []string) {
	fmt.Println("Choose a Google Photos-related operation like \"picsync googlephotos login\" or \"picsync googlephotos sync\"")
	fmt.Println("(try --help to see options)")
	os.Exit(1)
}

func runGooglephotosLogin(cmd *cobra.Command, args []string) {
	consumerKey := viper.GetString("googlephotos_api_key")
	if consumerKey == "" {
		panic("Must provide a Google Photos API key")
	}
	consumerSecret := viper.GetString("googlephotos_api_secret")
	if consumerSecret == "" {
		panic("Must provide a Google Photos API secret")
	}
	accessAuth, err := googlephotos.Login(consumerKey, consumerSecret)
	if err != nil {
		fmt.Printf("Login error: %v", err)
	}
	auth := &googlephotos.GooglephotosAuth{
		Access: *accessAuth,
	}

	toWrite := fmt.Sprintf(
		"# Keep this file confidential.\n"+
			"# If you lose it, de-authorize nixplay-sync from your Google Photos account and repeat 'picsync googlephotos login'\n"+
			"googlephotos:\n"+
			"  access:\n"+
			"    token_type: \"%s\"\n"+
			"    access_token: \"%s\"\n"+
			"    refresh_token: \"%s\"\n"+
			"    expiry: \"%s\"\n",
		auth.Access.TokenType,
		auth.Access.AccessToken,
		auth.Access.RefreshToken,
		auth.Access.Expiry.Format(time.RFC3339),
	)
	writeLoginOut(toWrite)
}

func newGooglePhotosClient() (*http.Client, error) {
	var err error

	consumerKey := viper.GetString("googlephotos_api_key")
	if consumerKey == "" {
		return nil, fmt.Errorf("must provide a Google Photos API key")
	}
	consumerSecret := viper.GetString("googlephotos_api_secret")
	if consumerSecret == "" {
		return nil, fmt.Errorf("must provide a Google Photos API secret")
	}

	access := oauth2.Token{
		TokenType:    viper.GetString("googlephotos.access.token_type"),
		AccessToken:  viper.GetString("googlephotos.access.access_token"),
		RefreshToken: viper.GetString("googlephotos.access.refresh_token"),
	}

	expiryString := viper.GetString("googlephotos.access.expiry")
	if expiryString != "" {
		access.Expiry, err = time.Parse(time.RFC3339, expiryString)
		if err != nil {
			return nil, err
		}
	}
	fmt.Printf("Access: %v\n", access)

	c := googlephotos.Client(consumerKey, consumerSecret, context.Background(), &access)
	return c, nil
}

func runGooglephotosList(cmd *cobra.Command, args []string) {
	c, err := newGooglePhotosClient()
	if err != nil {
		panic(err)
	}

	if len(args) == 0 {
		albums, err := googlephotos.ListAlbums(c)
		if err != nil {
			panic(err)
		}
		for _, a := range albums {
			fmt.Printf("Album \"%s\":\n", a.Title)
			fmt.Printf("  ID: %s\n", a.Id)
			fmt.Printf("  Items: %d\n", a.MediaItemsCount)
			fmt.Printf("  Google Photos: %s\n", a.ProductUrl)
		}
		return
	}

	if len(args) == 1 {
		albumId := args[0]
		items, err := googlephotos.ListMediaItemsForAlbumId(c, albumId)
		if err != nil {
			panic(err)
		}
		for _, item := range items {
			fmt.Printf("Media Item \"%s\":\n", item.Filename)
			fmt.Printf("  ID: %s\n", item.Id)
			fmt.Printf("  Description: %s\n", item.Description)
			fmt.Printf("  Google Photos: %s\n", item.ProductUrl)
			fmt.Printf("  Raw: %s\n", item.BaseUrl)
		}
		return
	}
	panic("Unexpected number of args")
}
