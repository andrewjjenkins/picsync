package main

import (
	"fmt"

	"github.com/andrewjjenkins/picsync/pkg/cache"
	"github.com/spf13/cobra"
)

var (
	cacheCmd = &cobra.Command{
		Use:   "cache",
		Short: "(advanced) operate on the metadata cache",
	}

	cacheStatusCmd = &cobra.Command{
		Use:   "status",
		Short: "Report status of the cache",
		Run:   runCacheStatus,
	}

	cacheFilename = ""
)

func init() {
	cacheCmd.AddCommand(cacheStatusCmd)

	rootCmd.AddCommand(cacheCmd)

	rootCmd.PersistentFlags().StringVar(
		&cacheFilename,
		"cache",
		"picsync-metadata-cache.db",
		"Path to cache file",
	)
}

func runCacheStatus(cmd *cobra.Command, args []string) {
	cache, err := cache.New(promReg, cacheFilename)
	if err != nil {
		panic(err)
	}

	status, err := cache.Status()
	if err != nil {
		panic(err)
	}
	fmt.Printf("Cache status:\n"+
		"Google Photos Valid Entries: %d\n"+
		"Nixplay Valid Entries: %d\n",
		status.GooglePhotosValidRows,
		status.NixplayValidRows,
	)
}
