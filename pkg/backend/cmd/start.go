package cmd

import (
	"github.com/homenoc/jpnic-admin-daemon/pkg/api/core"
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
		if err != nil || confPath == "" {
			log.Printf("getting environment")
			err = config.GetEnvConfig()
			if err != nil {
				log.Fatal("getting os.env config", err)
			}
		} else {
			if config.GetConfig(confPath) != nil {
				log.Fatal("getting config", err)
			}
		}

		err = config.ParseDatabase()
		if err != nil {
			log.Fatalf("getting database error: %v", err)
		}

		err = config.GetCA()
		if err != nil {
			log.Fatalf("getting CA Cert error: %v", err)
		}
		//log.Println("Database:", config.ConfDatabase.Driver, config.ConfDatabase.Option)

		core.Start()

		log.Println("end")
	},
}

func init() {
	rootCmd.AddCommand(startCmd)
	startCmd.PersistentFlags().StringP("config", "c", "", "config path")
	startCmd.PersistentFlags().StringP("template", "t", "", "config path")
}
