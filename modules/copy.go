package modules

import (
	"errors"

	"golang.org/x/crypto/ssh"
)

type CopyModule struct {
	src string
	dst string
}

func NewCopy(args []string) (Module, error) {
	if len(args) < 2 {
		return nil, errors.New("not enough arguments")
	}
	return &CopyModule{args[0], args[1]}, nil
}

func (cm *CopyModule) RunModule(client *ssh.Client, host string) (string, error) {
	return "", nil
}
