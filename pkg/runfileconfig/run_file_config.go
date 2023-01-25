package runfileconfig

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const defaultConfigFileName = "config.json"
const defaultConfigDirPath = ".lotta"

var forbiddenCommandNames = []string{"add", "help", "completion"}

type Commander interface {
	AddCommand(name string, cmd Command) error
	DeleteCommand(name string) error
}

type RunFileConfig struct {
	filepath     string
	UseSystemEnv bool
	Commands     map[string]Command `json:"commands"`
}

func LoadRunFileConfig() (*RunFileConfig, error) {
	config := RunFileConfig{}
	if err := config.loadConfigFile(); err != nil {
		return nil, err
	}

	return &config, nil
}

func (config *RunFileConfig) AddCommand(name string, cmd Command) error {
	err := config.validateConfig()
	if err != nil {
		return err
	}

	name, err = config.validateCommandName(name)
	if err != nil {
		return err
	}

	err = cmd.validate()
	if err != nil {
		return err
	}

	config.Commands[name] = cmd

	err = config.save()
	if err != nil {
		return err
	}
	return nil
}

func (config *RunFileConfig) DeleteCommand(name string) error {
	if _, ok := config.Commands[name]; !ok {
		return fmt.Errorf("no such command: %s", name)
	}

	delete(config.Commands, name)
	err := config.save()
	if err != nil {
		return err
	}
	return nil
}

func (config *RunFileConfig) save() error {
	content, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("error in marshaling config: %w", err)
	}

	err = os.WriteFile(config.filepath, content, 0644)
	if err != nil {
		return fmt.Errorf("error in writing config file: %w", err)
	}

	return nil
}

func (config *RunFileConfig) GetEnv() []string {
	return nil
}

func (config *RunFileConfig) loadConfigFile() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("error in reading home directory: %w", err)
	}
	config.filepath = filepath.Join(home, defaultConfigDirPath, defaultConfigFileName)

	// If config file doesn't exist, create it with emtpy commands
	if _, err := os.Stat(config.filepath); os.IsNotExist(err) {
		err = os.MkdirAll(filepath.Join(home, defaultConfigDirPath), 0700)
		if err != nil {
			return fmt.Errorf("error in config dir path: %w", err)
		}

		content, err := json.Marshal(config)
		if err != nil {
			return fmt.Errorf("error in marshaling config: %w", err)
		}

		err = os.WriteFile(config.filepath, content, 0644)
		if err != nil {
			return fmt.Errorf("error in writing config file: %w", err)
		}

		return nil
	}

	bytes, err := os.ReadFile(config.filepath)
	if err != nil {
		return fmt.Errorf("error in reading the config file: %w", err)
	}

	err = json.Unmarshal(bytes, &config)
	if err != nil {
		return err
	}

	return nil
}

func (config *RunFileConfig) validateConfig() error {
	if config.Commands == nil {
		config.Commands = make(map[string]Command)
		return nil
	}

	return nil
}

func (config *RunFileConfig) validateCommandName(name string) (string, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return "", fmt.Errorf("Command name can't be empty!")
	}

	for _, forbiddenName := range forbiddenCommandNames {
		if name == forbiddenName {
			return "", fmt.Errorf("Command with %s name is forbidden", name)
		}
	}
	return name, nil
}
