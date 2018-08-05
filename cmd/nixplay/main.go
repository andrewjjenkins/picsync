package main

import (
	"encoding/base64"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/andrewjjenkins/nixplay/pkg/nixplay"
	"github.com/andrewjjenkins/nixplay/pkg/smugmug"
)

var (
	rootCmd = &cobra.Command{
		Use:   "picsync",
		Short: "sync pictures from SmugMux to nixplay",
		Run:   run,
	}
)

func init() {
	viper.SetEnvPrefix("picsync")
	viper.BindEnv("nixplay_password")
	viper.BindEnv("nixplay_username")
	viper.BindEnv("smugmug_username")
	viper.BindEnv("smugmug_password")
}

func doNixplay() {
	username := viper.GetString("nixplay_username")
	if username == "" {
		panic("Must provide a nixplay username")
	}
	password := viper.GetString("nixplay_password")
	if password == "" {
		panic("Must provide a nixplay password")
	}
	c, err := nixplay.Login(username, password)
	if err != nil {
		fmt.Printf("Login error: %v", err)
	}
	albums, err := nixplay.GetAlbums(c)
	if err != nil {
		fmt.Printf("Fetch albums error: %v", err)
	}
	fmt.Printf("Albums: %v\n", albums)
	if len(albums) > 0 {
		album := albums[0]
		photos, err := nixplay.GetPhotos(c, album.ID)
		if err != nil {
			fmt.Printf("Error fetching photos: %v\n", err)
		}
		fmt.Printf("Photos for %s: %v\n", album.Title, photos)
	}
}

func doSmugmug() {
	username := viper.GetString("smugmug_username")
	if username == "" {
		panic("Must provide a smugmug username")
	}
	password := viper.GetString("smugmug_password")
	if password == "" {
		panic("Must provide a smugmug password")
	}
	_, err := smugmug.Login(username, password)
	if err != nil {
		fmt.Printf("Login error: %v", err)
	}
}

func run(cmd *cobra.Command, args []string) {
	//doNixplay()
	doSmugmug()
}

func main() {
	rootCmd.Execute()
}
