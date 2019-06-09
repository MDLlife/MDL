/*
cli is a command line client for interacting with a MDL node and offline wallet management
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

	mdlCLI, err := cli.NewCLI(cfg)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if err := mdlCLI.Execute(); err != nil {
		os.Exit(1)
	}
}
