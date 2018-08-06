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
		Short: "sync pictures from SmugMux to nixplay",
		Run:   run,
	}

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
	rootCmd.AddCommand(syncSmugmugToNixplay)
}

func run(cmd *cobra.Command, args []string) {
	fmt.Printf("Choose a command (try --help to see options)")
	os.Exit(1)
}

func main() {
	rootCmd.Execute()
}
