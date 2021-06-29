package main

import (
	"fmt"
	"strings"
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
	clienConfig *ssh.ClientConfig
	inventory   []string
	outputMutex sync.Mutex
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

	clienConfig = &ssh.ClientConfig{
		User:            viper.GetString("auth.user"),
		Auth:            authMethods,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         viper.GetDuration("timeout")}
}

func doStuff(commands, hosts []string, wg *sync.WaitGroup) {
	defer wg.Done()
	port := viper.GetInt("port")

	for i, step := 0, viper.GetInt("threads"); i < len(hosts); i += step {
		client, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", hosts[i], port), clienConfig)
		if err != nil {
			log.Error(err)
			continue
		}
		for _, cmd := range commands {
			session, err := client.NewSession()
			if err != nil {
				log.Error(err)
				continue
			}

			output, err := session.CombinedOutput(cmd)
			outputMutex.Lock()
			if err != nil {
				fmt.Printf("\n=== %s ===\n%s\n%s\n", hosts[i], strings.TrimSpace(string(output)+err.Error()), strings.Repeat("=", len(hosts[i])+8))
			} else {
				fmt.Printf("\n=== %s ===\n%s\n%s\n", hosts[i], strings.TrimSpace(string(output)), strings.Repeat("=", len(hosts[i])+8))
			}
			outputMutex.Unlock()
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
	wg.Add(threads)
	for i := 0; i < threads; i++ {
		if i >= len(inventory) {
			break
		}
		go doStuff(args, inventory[i:], &wg)
	}
	wg.Wait()
}
