package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "backend",
	Short: "web system backend",
	Long:  "Source Code: https://github.com/homenoc/dsbd-backend",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.controller.yaml)")

	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
