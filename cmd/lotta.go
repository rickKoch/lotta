package cmd

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/fatih/color"
	"github.com/rickKoch/lotta/pkg/runexec"
	"github.com/rickKoch/lotta/pkg/runfileconfig"
	config "github.com/rickKoch/lotta/pkg/runfileconfig"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var lottaConfigPath string
var runningCMD = color.New(color.FgHiGreen, color.Bold)

var RootCmd = &cobra.Command{
	Use:   "lotta",
	Short: "Lotta Command Palette",
	Long: `
	 __      ______  ______ ______ ______
	/\ \    /\  __ \/\__  _/\__  _/\  __ \
	\ \ \___\ \ \/\ \/_/\ \\/_/\ \\ \  __ \
	 \ \_____\ \_____\ \ \_\  \ \_\\ \_\ \_\
	  \/_____/\/_____/  \/_/   \/_/ \/_/\/_/
	========================================
	             Command Palette
	`,
}

func Execute(version string) {
	RootCmd.Version = version

	cobra.OnInitialize(initConfig)

	if err := RootCmd.Execute(); err != nil {
		fmt.Printf("There's been an error: %v", err)
		os.Exit(-1)
	}
}

func initConfig() {
	viper.SetEnvPrefix("lotta")
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.AutomaticEnv()
}

func init() {
	config, err := config.LoadRunFileConfig()
	if err != nil {
		fmt.Printf("There's been an error: %v", err)
		os.Exit(1)
	}

	for cmdName, command := range config.Commands {
		createCommand(config, cmdName, command)
	}
}

func createCommand(config *runfileconfig.RunFileConfig, name string, command config.Command) {
	customCmd := &cobra.Command{
		Use:   name,
		Short: command.Description,
		Run: func(cmd *cobra.Command, args []string) {
			cmdProcess, err := runexec.GetCmdProcess(config, cmd.Use, args)
			if err != nil {
				log.Fatal(err)
			}

			sigCh := make(chan os.Signal, 1)
			setupShutdownNotify(sigCh)

			runningCMD.Printf("⌛  %s\n", cmdProcess.Command.String())
			if err := cmdProcess.Command.Start(); err != nil {
				fmt.Printf("There's been an error: %v", err)
				os.Exit(1)
			}

			go func() {
				if err := cmdProcess.Command.Wait(); err != nil {
					cmdProcess.CommandErr = err
				}
				sigCh <- os.Interrupt
			}()
			<-sigCh

			if cmdProcess.Command.ProcessState == nil || !cmdProcess.Command.ProcessState.Exited() {
				err = cmdProcess.Command.Process.Kill()
				if err != nil {
					fmt.Fprintf(os.Stderr, fmt.Sprintf("Error exiting Command: %s", err))
				}
			}

			if cmdProcess.CommandErr != nil {
				os.Exit(1)
			}
		},
	}
	for _, flag := range command.Flags {
		customCmd.Flags().StringVar(&flag.Value, flag.Name, flag.Value, "")
		if flag.Required {
			customCmd.MarkFlagRequired(flag.Name)
		}
	}
	RootCmd.AddCommand(customCmd)
}

func setupShutdownNotify(sigCh chan os.Signal) {
	signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT)
}
