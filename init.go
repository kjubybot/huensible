package main

import (
	"errors"
	"huensible/modules"
	"io/ioutil"
	"regexp"

	"github.com/spf13/viper"
	"golang.org/x/crypto/ssh"
)

func parseAuthMethods() ([]ssh.AuthMethod, error) {
	pass := viper.GetString("auth.password")
	pkey := viper.GetString("auth.pkey")

	if pass == "" && pkey == "" {
		return nil, nil
	}

	auth := make([]ssh.AuthMethod, 0, 2)

	if pkey != "" {
		keyRaw, err := ioutil.ReadFile(pkey)
		if err != nil {
			return nil, err
		}

		key, err := ssh.ParsePrivateKey(keyRaw)
		if err != nil {
			return nil, err
		}

		auth = append(auth, ssh.PublicKeys(key))
	}

	if pass != "" {
		auth = append(auth, ssh.Password(pass))
	}
	return auth, nil
}

func parseInventory(inventoryFile string) ([]string, error) {
	ipRegex := regexp.MustCompile(`\d+\.\d+\.\d+\.\d+(-\d+\.\d+\.\d+\.\d+)?`)
	invViper := viper.New()
	invViper.SetConfigFile(inventoryFile)

	if err := invViper.ReadInConfig(); err != nil {
		return nil, err
	}

	hosts := invViper.GetStringSlice("hosts")
	if len(hosts) == 0 {
		return nil, errors.New("no hosts in inventory")
	}

	parsedHosts := make([]string, 0, len(hosts))
	for _, host := range hosts {
		if ipRegex.Match([]byte(host)) {
			parsedHosts = append(parsedHosts, parseIps(host)...)
		} else {
			parsedHosts = append(parsedHosts, host)
		}
	}

	return parsedHosts, nil
}

func parseModule(moduleFlag string) (modules.ModuleFactory, error) {
	if mod, ok := modules.Modules[moduleFlag]; ok {
		return mod, nil
	} else {
		return nil, errors.New("module not found")
	}
}
