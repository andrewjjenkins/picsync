package main

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/andrewjjenkins/nixplay/pkg/nixplay"
)

var (
	rootCmd = &cobra.Command{
		Use:   "nixplay",
		Short: "sync pictures from SmugMux to nixplay",
		Run:   run,
	}
)

func init() {
	viper.SetEnvPrefix("nixplay")
	viper.BindEnv("password")
	viper.BindEnv("username")
}

func run(cmd *cobra.Command, args []string) {
	username := viper.GetString("USERNAME")
	if username == "" {
		panic("Must provide a username")
	}
	password := viper.GetString("password")
	if password == "" {
		panic("Must provide a password")
	}
	c, err := nixplay.Login(username, password)
	if err != nil {
		fmt.Printf("Login error: %v", err)
	}
	nixplay.GetConfig(c)
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

func main() {
	rootCmd.Execute()
}
