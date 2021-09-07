package modules

import (
	"fmt"
	"strings"

	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
)

type CommandModule struct {
	args []string
}

func NewCommand(args []string) (Module, error) {
	return &CommandModule{args}, nil
}

func (cm *CommandModule) RunModule(client *ssh.Client, host string) (string, error) {
	var result strings.Builder

	for _, cmd := range cm.args {
		session, err := client.NewSession()
		if err != nil {
			log.Error(err)
			return "", err
		}

		output, err := session.CombinedOutput(cmd)
		if err != nil {
			fmt.Fprintf(&result, "\n=== %s ===\n%s\n%s\n", host, strings.TrimSpace(string(output)+err.Error()), strings.Repeat("=", len(host)+8))
		} else {
			fmt.Fprintf(&result, "\n=== %s ===\n%s\n%s\n", host, strings.TrimSpace(string(output)), strings.Repeat("=", len(host)+8))
		}
	}
	return result.String(), nil
}
