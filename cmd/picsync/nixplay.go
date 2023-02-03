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

	nixplayDeleteCmd = &cobra.Command{
		Use:   "delete",
		Short: "Delete albums or photos",
	}

	nixplayDeleteAlbumCmd = &cobra.Command{
		Use:   "album <albumName>",
		Short: "Delete all albums named <albumName>",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return fmt.Errorf("must specify album to delete")
			}
			albumName := args[0]
			if albumName == "" {
				return fmt.Errorf("must specify album to delete")
			}
			runNixplayDeleteAlbum(albumName)
			return nil
		},
	}

	allowDeleteMultiple bool
)

func init() {
	nixplayListCmd.PersistentFlags().BoolVar(
		&updateCache,
		"update-cache",
		false,
		"Also update the cache when listing (temporarily downloads each image)",
	)
	nixplayDeleteAlbumCmd.PersistentFlags().BoolVar(
		&allowDeleteMultiple,
		"delete-multiple",
		false,
		"If there are multiple albums with the same name, delete them all instead of quitting",
	)

	nixplayCmd.AddCommand(nixplayListCmd)
	nixplayDeleteCmd.AddCommand(nixplayDeleteAlbumCmd)
	nixplayCmd.AddCommand(nixplayDeleteCmd)
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
		fmt.Printf("  ID: %d\n", a.ID)
	}
}

func runNixplayListAlbum(albumName string) {
	npClient := getNixplayClientOrExit()

	npAlbums, err := npClient.GetAlbumsByName(albumName)
	if err != nil {
		panic(err)
	}
	for _, npAlbum := range npAlbums {
		page := 1
		limit := 100
		fmt.Printf("Photos for album %s (%d)\n", npAlbum.Title, npAlbum.ID)
		for {
			npPhotos, err := npClient.GetPhotos(npAlbum.ID, page, limit)
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

			for i, p := range npPhotos {
				fmt.Printf("Nixplay Photo %d:\n", i+((page-1)*limit))
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

			if len(npPhotos) < limit {
				break
			}
			page++
		}
	}
}

func runNixplayDeleteAlbum(albumName string) {
	npClient := getNixplayClientOrExit()

	deletedCount, err := npClient.DeleteAlbumsByName(albumName, allowDeleteMultiple)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Deleted %d albums named %s\n", deletedCount, albumName)
}
