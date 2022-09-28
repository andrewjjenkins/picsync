package main

import (
	"fmt"
	"time"

	"github.com/andrewjjenkins/picsync/pkg/cache"
	"github.com/spf13/cobra"
)

var (
	nixplayCmd = &cobra.Command{
		Use:   "nixplay",
		Short: "Nixplay-specific operations",
	}

	nixplayListCmd = &cobra.Command{
		Use:   "list [<albumName>]",
		Short: "List albums or photos in a specific album",
		Run:   runNixplayList,
	}
)

func init() {
	nixplayListCmd.PersistentFlags().BoolVar(
		&updateCache,
		"update-cache",
		false,
		"Also update the cache when listing (temporarily downloads each image)",
	)

	nixplayCmd.AddCommand(nixplayListCmd)
	rootCmd.AddCommand(nixplayCmd)
}

func runNixplayList(cmd *cobra.Command, args []string) {
	if len(args) > 1 {
		panic(fmt.Errorf("only one argument allowed (album name)"))
	}
	if len(args) == 1 {
		albumName := args[0]
		runNixplayListAlbum(albumName)
		return
	}
	runNixplayListAlbums()
}

func runNixplayListAlbums() {
	npClient := getNixplayClientOrExit()

	npAlbums, err := npClient.GetAlbums()
	if err != nil {
		panic(err)
	}
	for _, a := range npAlbums {
		fmt.Printf("Nixplay album %s:\n", a.Title)
		fmt.Printf("  Photos: %d\n", a.PhotoCount)
		fmt.Printf("  Published: %t\n", a.Published)
	}
}

func runNixplayListAlbum(albumName string) {
	npClient := getNixplayClientOrExit()

	npAlbum, err := npClient.GetAlbumByName(albumName)
	if err != nil {
		panic(err)
	}
	npPhotos, err := npClient.GetPhotos(npAlbum.ID)
	if err != nil {
		panic(err)
	}

	var c cache.Cache
	if updateCache {
		c, err = cache.New(promReg, cacheFilename)
		if err != nil {
			panic(err)
		}
	}

	fmt.Printf("Photos for album %s (%d)\n", npAlbum.Title, npAlbum.ID)
	for i, p := range npPhotos {
		fmt.Printf("Nixplay Photo %d:\n", i)
		fmt.Printf("  Filename: %s\n", p.Filename)
		fmt.Printf("  Date: %s\n", p.SortDate)
		fmt.Printf("  URL: %s\n", p.URL)
		fmt.Printf("  MD5: %s\n", p.Md5)
		if updateCache {
			timeNow := time.Now()
			err := c.UpsertNixplay(&cache.NixplayData{
				NixplayId:   p.ID,
				URL:         p.URL,
				Filename:    p.Filename,
				SortDate:    p.SortDate,
				Md5:         p.Md5,
				LastUpdated: timeNow,
				LastUsed:    timeNow,
			})
			if err != nil {
				panic(err)
			}
		}
	}
}
