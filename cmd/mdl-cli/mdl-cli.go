/*
cli is a command line client for interacting with a skycoin node and offline wallet management
*/
package main

import (
	"fmt"
    "github.com/MDLlife/MDL/src/cli"
    "github.com/MDLlife/MDL/src/util/logging"
    "os"

	"github.com/sirupsen/logrus"

)

func main() {
	logging.SetLevel(logrus.WarnLevel)

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
