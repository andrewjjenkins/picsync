package main

import (
	"fmt"

	"github.com/andrewjjenkins/picsync/pkg/nixplay"
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
	nixplayCmd.AddCommand(nixplayListCmd)
	rootCmd.AddCommand(nixplayCmd)
}

func runNixplayList(cmd *cobra.Command, args []string) {
	if len(args) > 1 {
		panic(fmt.Errorf("only one argument allowed (album name)"))
	}
	if len(args) == 1 {
		panic(fmt.Errorf("not implemented"))
	}
	runNixplayListAlbums()
}

func runNixplayListAlbums() {
	npClient := getNixplayClientOrExit()

	npAlbums, err := nixplay.GetAlbums(npClient)
	if err != nil {
		panic(err)
	}
	for _, a := range npAlbums {
		fmt.Printf("Nixplay album %s:\n", a.Title)
		fmt.Printf("  Photos: %d\n", a.PhotoCount)
		fmt.Printf("  Published: %t\n", a.Published)
	}
}
