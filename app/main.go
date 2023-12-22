package main

import (
	"flag"
	"fmt"

	"tianxian.com/tee-signer-core/common"
	"tianxian.com/tee-signer-core/config"
	"tianxian.com/tee-signer-core/jobs"
	"tianxian.com/tee-signer-core/log"
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
		config.LoadConfig(*configFile)
	} else {
		// config file path
		configPath := fmt.Sprintf("./config-%s.toml", *env)
		log.Entry.Infof("config file: %s", configPath)
		config.LoadConfig(configPath)
	}

	tasks := []func(chan struct{}){common.Flush}
	tasks = append(tasks, jobs.GetJobs()...)
	common.DoLoopJobs(tasks...)
}
