package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/andrewjjenkins/picsync/pkg/cache"
	"github.com/andrewjjenkins/picsync/pkg/googlephotos"
	"github.com/spf13/cobra"
	"golang.org/x/oauth2"
)

var (
	googlephotosCmd = &cobra.Command{
		Use:   "googlephotos",
		Short: "Interact with Google Photos",
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

	listShared = false
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

	googlephotosList.PersistentFlags().BoolVar(
		&updateCache,
		"update-cache",
		false,
		"Also update the cache when listing (temporarily downloads each image)",
	)

	googlephotosList.PersistentFlags().BoolVar(
		&listShared,
		"shared",
		false,
		"List albums shared with you",
	)

	googlephotosCmd.AddCommand(googlephotosList)

	rootCmd.AddCommand(googlephotosCmd)
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
	c := googlephotos.Client(consumerKey, consumerSecret, context.Background(), &access)
	return c, nil
}

func getGooglephotoClientOrExit() (c *http.Client) {
	c, err := newGooglePhotosClient()
	if err != nil {
		fmt.Printf("Google Photos login error: %v", err)
		os.Exit(1)
	}
	return c
}

func runGooglephotosList(cmd *cobra.Command, args []string) {
	c := getGooglephotoClientOrExit()

	if len(args) == 0 {
		var albums []*googlephotos.Album
		var err error
		if listShared {
			albums, err = googlephotos.ListSharedAlbums(c)
		} else {
			albums, err = googlephotos.ListAlbums(c)
		}
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
		if !updateCache {
			var nextPageToken string
			for ok := true; ok; ok = (nextPageToken != "") {
				resp, err := googlephotos.ListMediaItemsForAlbumId(c, albumId, nextPageToken)
				if err != nil {
					panic(err)
				}
				nextPageToken = resp.NextPageToken
				for _, item := range resp.MediaItems {
					fmt.Printf("Media Item \"%s\":\n", item.Filename)
					fmt.Printf("  Google Photos ID: %s\n", item.Id)
					fmt.Printf("  Description: %s\n", item.Description)
					fmt.Printf("  Google Photos: %s\n", item.ProductUrl)
					fmt.Printf("  Raw: %s\n", item.BaseUrl)
				}
			}
			return
		}
		runGooglephotosListUpdateCache(c, albumId)
		return
	}
	panic("Unexpected number of args")
}

func runGooglephotosListUpdateCache(client *http.Client, albumId string) {
	var updatedCount int64
	updateCallback := func(cached *googlephotos.CachedMediaItem) {
		updatedCount++
		fmt.Printf("Updated %d:\n", updatedCount)
		fmt.Printf("  Google Photos ID: %s\n", cached.MediaItem.Id)
		fmt.Printf("  Description: %s\n", cached.MediaItem.Description)
		fmt.Printf("  Google Photos: %s\n", cached.MediaItem.ProductUrl)
		fmt.Printf("  Raw: %s\n", cached.MediaItem.BaseUrl)
		fmt.Printf("  Cache ID/Md5/Sha256: %d/%s/%s\n",
			cached.CacheId, cached.Md5, cached.Sha256)
	}

	c, err := cache.New(promReg)
	if err != nil {
		panic(err)
	}

	var nextPageToken string
	for ok := true; ok; ok = (nextPageToken != "") {
		res, err := googlephotos.UpdateCacheForAlbumId(client, c, albumId, nextPageToken, updateCallback)
		if err != nil {
			panic(err)
		}
		nextPageToken = res.NextPageToken
	}
}
