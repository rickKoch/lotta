package runexec

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"text/template"

	"github.com/rickKoch/lotta/pkg/runfileconfig"
)

type CmdProcess struct {
	Command      *exec.Cmd
	CommandErr   error
	OutputWriter io.Writer
	ErrorWriter  io.Writer
}

func GetCmdProcess(config *runfileconfig.RunFileConfig, cmdName string, args []string) (*CmdProcess, error) {
	cmd, err := getCommand(config, cmdName, args)
	if err != nil {
		return nil, err
	}

	return &CmdProcess{
		Command: cmd,
	}, nil
}

func getCommand(config *runfileconfig.RunFileConfig, cmdName string, args []string) (*exec.Cmd, error) {
	runFileConfigCmd, ok := config.Commands[cmdName]
	if !ok {
		return nil, fmt.Errorf("No such command: %s", cmdName)
	}

	flags := make(map[string]string)
	for _, flag := range runFileConfigCmd.Flags {
		flags[flag.Name] = flag.Value
	}

	t, err := template.New("exec").Parse(runFileConfigCmd.Exec)
	if err != nil {
		return nil, err
	}

	var buff bytes.Buffer
	err = t.Execute(&buff, flags)
	if err != nil {
		return nil, err
	}

	cmdExecArgs := strings.Split(buff.String(), " ")
	argCount := len(cmdExecArgs)

	if argCount == 0 {
		return nil, fmt.Errorf("No such command: %s", cmdName)
	}
	command := cmdExecArgs[0]

	cmdArgs := []string{}
	if argCount > 1 {
		cmdArgs = cmdExecArgs[1:]
	}
	cmdArgs = append(cmdArgs, args...)

	cmd := exec.Command(command, cmdArgs...)
	if config.UseSystemEnv {
		cmd.Env = os.Environ()
	}
	cmd.Env = append(cmd.Env, config.GetEnv()...)

	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin

	return cmd, nil
}
