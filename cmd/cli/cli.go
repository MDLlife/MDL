/*
cli is a command line client for interacting with a mdl node and offline wallet management
*/
package main

import (
	"fmt"
	"os"

	"github.com/MDLlife/MDL/src/cli"
)

func main() {
	cfg, err := cli.LoadConfig()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	skyCLI, err := cli.NewCLI(cfg)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if err := skyCLI.Execute(); err != nil {
		os.Exit(1)
	}
}
