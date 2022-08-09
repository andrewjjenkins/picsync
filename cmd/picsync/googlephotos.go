package main

import (
	"fmt"
	"os"

	"github.com/andrewjjenkins/picsync/pkg/googlephotos"
	"github.com/spf13/cobra"
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
		auth.Access.Expiry,
	)
	writeLoginOut(toWrite)
}
