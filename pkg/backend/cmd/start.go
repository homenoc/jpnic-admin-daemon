package cmd

import (
	"github.com/homenoc/jpnic-admin-daemon/pkg/api"
	"github.com/homenoc/jpnic-admin-daemon/pkg/api/core/tool/config"
	"github.com/spf13/cobra"
	"log"
)

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "start controller",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		confPath, err := cmd.Flags().GetString("config")
		if err != nil {
			log.Fatalf("could not greet: %v", err)
		}

		if config.GetConfig(confPath) != nil {
			log.Fatal("getting config", err)
		}

		api.RestAPI()

		log.Println("end")
	},
}

func init() {
	rootCmd.AddCommand(startCmd)
	startCmd.PersistentFlags().StringP("config", "c", "", "config path")
	startCmd.PersistentFlags().StringP("template", "t", "", "config path")
}
