package cmd

import (
	"fmt"
	"os"

	"github.com/rickKoch/lotta/pkg/gui"
	config "github.com/rickKoch/lotta/pkg/runfileconfig"
	"github.com/spf13/cobra"
)

var AddCmd = &cobra.Command{
	Use:     "add",
	Short:   "Add/update command to the Command Palette",
	Example: " add",
	Run: func(cmd *cobra.Command, args []string) {
		runFileConfig, err := config.LoadRunFileConfig()
		if err != nil {
			fmt.Printf("%v\n", err)
			os.Exit(1)
		}

		if err := gui.Run(runFileConfig); err != nil {
			fmt.Printf("There's been an error: %v", err)
			os.Exit(1)
		}
	},
}

func init() {
	RootCmd.AddCommand(AddCmd)
}
