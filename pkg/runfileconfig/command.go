package runfileconfig

import (
	"fmt"
	"strings"
)

type Command struct {
	Description string  `json:"description"`
	Exec        string  `json:"exec"`
	Flags       []*Flag `json:"flags"`
}

type Flag struct {
	Name     string `json:"name"`
	Value    string `json:"value"`
	Required bool   `json:"required"`
}

func (cmd *Command) validate() error {
	cmd.Exec = strings.TrimSpace(cmd.Exec)
	if cmd.Exec == "" {
		return fmt.Errorf("Command can't be empty!")
	}

	return nil
}
