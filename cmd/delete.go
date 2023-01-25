package cmd

import (
	"fmt"
	"os"

	config "github.com/rickKoch/lotta/pkg/runfileconfig"
	"github.com/spf13/cobra"
)

var deleteCmdName string

var DeleteCmd = &cobra.Command{
	Use:     "delete",
	Short:   "Delete command from the Command Palette",
	Example: "lotta delete --name example",
	Run: func(cmd *cobra.Command, args []string) {
		runFileConfig, err := config.LoadRunFileConfig()
		if err != nil {
			fmt.Printf("%v\n", err)
			os.Exit(1)
		}

		if err := runFileConfig.DeleteCommand(deleteCmdName); err != nil {
			fmt.Printf("There's been an error: %v", err)
			os.Exit(1)
		}
	},
}

func init() {
	DeleteCmd.Flags().StringVarP(&deleteCmdName, "name", "", "", "Delete command by name")
	DeleteCmd.MarkFlagRequired("name")
	RootCmd.AddCommand(DeleteCmd)
}
