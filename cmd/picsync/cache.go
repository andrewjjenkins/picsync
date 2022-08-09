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
)

func init() {
	cacheCmd.AddCommand(cacheStatusCmd)

	rootCmd.AddCommand(cacheCmd)
}

func runCacheStatus(cmd *cobra.Command, args []string) {
	db, err := cache.Open()
	if err != nil {
		panic(err)
	}

	status, err := cache.Status(db)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Cache status:\n%v\n", status)
}
