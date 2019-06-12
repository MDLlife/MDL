package cli

import (
	"fmt"
	"reflect"

	"github.com/spf13/cobra"
)

func versionCmd() *cobra.Command {
	versionCmd := &cobra.Command{
		Use:          "version",
		Short:        "List the current version of MDL components",
		Args:         cobra.NoArgs,
		SilenceUsage: true,
		RunE: func(c *cobra.Command, _ []string) error {
			var ver = struct {
				MDL    string `json:"mdl"`
				Cli    string `json:"cli"`
				RPC    string `json:"rpc"`
				Wallet string `json:"wallet"`
			}{
				Version,
				Version,
				Version,
				Version,
			}

			jsonOutput, err := c.Flags().GetBool("json")
			if err != nil {
				return err
			}
			if jsonOutput {
				return printJSON(ver)
			}

			v := reflect.ValueOf(ver)
			t := reflect.TypeOf(ver)
			for i := 0; i < v.NumField(); i++ {
				fmt.Printf("%s:%v\n", t.Field(i).Tag.Get("json"), v.Field(i).Interface())
			}

			return nil
		},
	}

	versionCmd.Flags().BoolP("json", "j", false, "Returns the results in JSON format")

	return versionCmd
}
