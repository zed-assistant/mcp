package main

import "flag"

type commandArgs struct {
	configPath string
}

func getFlags() *commandArgs {
	args := &commandArgs{}

	flag.StringVar(&args.configPath, "config", "config.yml", "path to the config file")
	flag.Parse()

	return args
}
