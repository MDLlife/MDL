package cli

import (
	cobra "github.com/spf13/cobra"

	"github.com/MDLlife/MDL/src/api"
)

// StatusResult is printed by cli status command
type StatusResult struct {
	Status api.HealthResponse `json:"status"`
	Config ConfigStatus       `json:"cli_config"`
}

// ConfigStatus contains the configuration parameters loaded by the cli
type ConfigStatus struct {
	RPCAddress string `json:"webrpc_address"`
}

func statusCmd() *cobra.Command {
	return &cobra.Command{
		Use:                   "status",
		Short:                 "Check the status of current mdl node",
		DisableFlagsInUseLine: true,
		SilenceUsage:          true,
		Args:                  cobra.NoArgs,
		RunE: func(c *cobra.Command, args []string) error {
			status, err := apiClient.Health()
			if err != nil {
				return err
			}

			return printJSON(StatusResult{
				Status: *status,
				Config: ConfigStatus{
					RPCAddress: cliConfig.RPCAddress,
				},
			})
		},
	}
}

func showConfigCmd() *cobra.Command {
	return &cobra.Command{
		Use:                   "showConfig",
		Short:                 "Show cli configuration",
		DisableFlagsInUseLine: true,
		RunE: func(c *cobra.Command, args []string) error {
			return printJSON(cliConfig)
		},
	}
}
