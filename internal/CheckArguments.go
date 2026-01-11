package internal

import (
	"forum/internal/database"
	"forum/internal/logger"
	"os"
	"strings"
)

func CheckArguments() {
	args := os.Args[1:]
	if len(args) == 0 {
		return // no flag passed
	}
	for _, flag := range args {
		flag = strings.TrimSpace(flag)
		switch flag {
		case "--debug", "-d":
			logger.Debug = true
			logger.Log("Debug messages are turned on", logger.InfoLevel)
		case "--seed", "-s":
			database.Seed = true
		case "--logs", "-l":
			logger.Enable = true
		case "--help", "-h":
			print(help())
			os.Exit(0)
		}
	}
}

func help() string {
	help := `Our server supports a few optional command-line flags to control runtime behavior:

	* --debug, -d
	  Enable debug logging. When set, the server will include debug-level messages in the logs.
	
	* --seed, -s
	  Populate the database with initial seed data on startup. Useful for development and testing environments.
	
	* --logs, -l
	  Enable general logging. When set, informational and error logs will be written to the console or configured log output.

	* --help, -h
	  Prints help message. 
	
	Flags can be passed when starting the application, for example:
	
	bash
	./forum-server --debug --seed`
	return help
}
