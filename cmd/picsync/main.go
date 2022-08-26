package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	viperLib "github.com/spf13/viper"
)

var (
	rootCmd = &cobra.Command{
		Use:   "picsync",
		Short: "sync pictures from Google Photos to nixplay",
		Run:   run,
	}

	viper *viperLib.Viper

	loginOut    string
	updateCache bool
)

func init() {
	viper = viperLib.New()
	viper.SetConfigName(".picsync-credentials")
	viper.AddConfigPath("./")
	viper.AddConfigPath("/etc/picsync-credentials/")
	err := viper.ReadInConfig()
	if err != nil {
		fmt.Printf("Error reading config file, skipping config file\n")
	}
}

func run(cmd *cobra.Command, args []string) {
	fmt.Printf("Choose a command (try --help to see options)")
	os.Exit(1)
}

func main() {
	rootCmd.Execute()
}
