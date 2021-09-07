package main

import (
	"fmt"
	"huensible/modules"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"golang.org/x/crypto/ssh"
)

type Result struct {
	Host   string
	Output string
}

var (
	clienConfig   *ssh.ClientConfig
	inventory     []string
	moduleFactory modules.ModuleFactory
	outputMutex   sync.Mutex
)

func init() {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./config/")

	viper.SetDefault("auth.user", "root")
	viper.SetDefault("threads", 4)
	viper.SetDefault("timeout", time.Duration(0))
	viper.SetDefault("port", 22)

	inventoryFile := pflag.StringP("inventory", "i", "", "inventory file to use")
	moduleFlag := pflag.StringP("module", "m", "command", "which module to use. Default is command")

	if err := viper.ReadInConfig(); err != nil {
		log.Panic(err)
	}

	if viper.GetInt("threads") <= 0 {
		log.Panic("threads cannot be zero or less")
	}

	if viper.GetInt("port") <= 0 {
		log.Panic("port cannot be zero or less")
	}

	pflag.Parse()

	if *inventoryFile == "" {
		log.Panic("no inventory found")
	}

	authMethods, err := parseAuthMethods()
	if err != nil {
		log.Panic(err)
	}

	inventory, err = parseInventory(*inventoryFile)
	if err != nil {
		log.Panic(err)
	}

	moduleFactory, err = parseModule(*moduleFlag)
	if err != nil {
		log.Panic(err)
	}

	clienConfig = &ssh.ClientConfig{
		User:            viper.GetString("auth.user"),
		Auth:            authMethods,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         viper.GetDuration("timeout")}
}

func doStuff(mod modules.Module, hosts []string, wg *sync.WaitGroup) {
	defer wg.Done()
	port := viper.GetInt("port")

	for i, step := 0, viper.GetInt("threads"); i < len(hosts); i += step {
		client, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", hosts[i], port), clienConfig)
		if err != nil {
			log.Error(err)
			continue
		}
		output, err := mod.RunModule(client, hosts[i])
		if err != nil {
			log.Error(err)
		} else {
			fmt.Print(output)
		}
	}
}

func main() {
	args := pflag.Args()
	if len(args) == 0 {
		log.Panic("no actions specified")
	}

	var wg sync.WaitGroup
	threads := viper.GetInt("threads")
	if threads > len(inventory) {
		threads = len(inventory)
	}

	mod, err := moduleFactory(args)
	if err != nil {
		log.Panic(err)
	}

	wg.Add(threads)
	for i := 0; i < threads; i++ {
		go doStuff(mod, inventory[i:], &wg)
	}
	wg.Wait()
}
