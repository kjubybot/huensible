package modules

import "golang.org/x/crypto/ssh"

type ModuleFactory func([]string) (Module, error)

var Modules map[string]ModuleFactory

func init() {
	Modules = make(map[string]ModuleFactory)
	Modules["command"] = NewCommand
	Modules["copy"] = NewCopy
}

type Module interface {
	RunModule(*ssh.Client, string) (string, error)
}
