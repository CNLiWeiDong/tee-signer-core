package main

import (
	"flag"
	"fmt"

	"blockchain.com/pump"
	"blockchain.com/pump/common"
	"blockchain.com/pump/log"
)

func main() {
	showVersion := flag.Bool("v", false, "show version")
	configFile := flag.String("c", "", "set the config file path")
	env := flag.String("env", "prod", "Operating environment: prod, test1, test2, test3")
	flag.Parse()

	if *showVersion {
		common.ShowVersion()
		return
	}

	if len(*configFile) > 0 {
		log.Entry.Infof("config file: %s", *configFile)
		pump.LoadConfig(*configFile)
	} else {
		// config file path
		configPath := fmt.Sprintf("./config-%s.toml", *env)
		log.Entry.Infof("config file: %s", configPath)
		pump.LoadConfig(configPath)
	}

	jobs := []func(chan struct{}){common.Flush}
	jobs = append(jobs, pump.GetJobs()...)
	common.DoLoopJobs(jobs...)
}
