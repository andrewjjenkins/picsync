package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	viperLib "github.com/spf13/viper"

	"github.com/andrewjjenkins/nixplay/pkg/nixplay"
	"github.com/andrewjjenkins/nixplay/pkg/smugmug"
)

var (
	rootCmd = &cobra.Command{
		Use:   "picsync",
		Short: "sync pictures from SmugMux to nixplay",
		Run:   run,
	}

	smugmugLogin = &cobra.Command{
		Use:   "login",
		Short: "Complete one-time log in to SmugMug (OAuth)",
		Run:   runSmugmugLogin,
	}

	smugmugLoginOut string

	viper *viperLib.Viper
)

func init() {
	viper = viperLib.New()
	viper.SetConfigName("picsync-config")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()
	if err != nil {
		fmt.Printf("Error reading config file, skipping config file\n")
	}

	viper.SetEnvPrefix("picsync")
	viper.BindEnv("nixplay_password")
	viper.BindEnv("nixplay_username")
	viper.BindEnv("smugmug_api_key")
	viper.BindEnv("smugmug_api_secret")

	smugmugLogin.PersistentFlags().StringVarP(&smugmugLoginOut, "outfile", "o", "", "Write token config out to file (like picsync-config.yaml)")

	rootCmd.AddCommand(smugmugLogin)
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
	auth := smugmug.AccessAuth{
		Token:          viper.GetString("smugmug.access.token"),
		Secret:         viper.GetString("smugmug.access.secret"),
		ConsumerKey:    viper.GetString("smugmug.consumer.token"),
		ConsumerSecret: viper.GetString("smugmug.consumer.secret"),
	}
	if auth.Token == "" || auth.Secret == "" {
		fmt.Printf("No smugmug auth; do \"picsync login -o picsync-config.yaml\"\n")
		os.Exit(1)
	}
	client, err := smugmug.Access(&auth)
	if err != nil {
		fmt.Printf("Smugmug auth failed (%v); "+
			"repeat \"picsync login -o picsync-config.yaml\"\n", err)
	}

	album, err := smugmug.GetAlbum(client, "SJT3DX")
	if err != nil {
		fmt.Printf("Getting album failed: %v\n", err)
	}
	fmt.Printf("%v\n", album)
}

func run(cmd *cobra.Command, args []string) {
	//doNixplay()
	doSmugmug()
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
	auth, err := smugmug.Login(consumerKey, consumerSecret)
	if err != nil {
		fmt.Printf("Login error: %v", err)
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
		auth.Token,
		auth.Secret,
		consumerKey,
		consumerSecret,
	)

	var outfile *os.File
	if smugmugLoginOut == "-" || smugmugLoginOut == "" {
		outfile = os.Stdout
	} else {
		outfile, err = os.Create(smugmugLoginOut)
		if err != nil {
			panic(err)
		}
		defer outfile.Close()
	}
	_, err = outfile.WriteString(toWrite)
	if err != nil {
		fmt.Printf("Error writing login info to %s: %v\n", smugmugLoginOut, err)
	}
	outfile.Sync()
}

func main() {
	rootCmd.Execute()
}
