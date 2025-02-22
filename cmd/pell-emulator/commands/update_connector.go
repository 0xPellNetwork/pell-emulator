package commands

import (
	"github.com/spf13/cobra"

	"github.com/0xPellNetwork/pell-emulator/cmd/pell-emulator/chainflags"
	"github.com/0xPellNetwork/pell-emulator/config"
	"github.com/0xPellNetwork/pell-emulator/internal/chains"
)

func init() {
	//chainflags.EmulatorFlagRPCURL.AddToCmdFlag(EmulatorUpdateConnectorCmd)
	//chainflags.EmulatorFlagWSURL.AddToCmdFlag(EmulatorUpdateConnectorCmd)
	//chainflags.EmulatorFlagAutoUpdateConnector.AddToCmdFlag(EmulatorUpdateConnectorCmd)
	//chainflags.EmulatorFlagDeployerKeyFile.AddToCmdFlag(EmulatorUpdateConnectorCmd)
}

var EmulatorUpdateConnectorCmd = &cobra.Command{
	Use:   "update-connector",
	Short: "update connector",
	Example: `
pell-emulator update-connector \
	--home <home-dir> \
	--rpc-url <rpc-url> \
	--ws-url <ws-url> \
	--deployer-key-file <deployer-key-file>

pell-emulator update-connector \
	--home <home-dir> \
	--rpc-url http://localhost:8545 \
	--ws-url ws://localhost:8545 \
	--deployer-key-file /path/to/deployer-key-file

`,
	RunE: func(cmd *cobra.Command, args []string) error {
		logger.Info("start update connector",
			"k", "v",
			"rpcURL", chainflags.EmulatorFlagRPCURL.GetValue(),
			"wsURL", chainflags.EmulatorFlagWSURL.GetValue(),
			"contractAddressFile", chainflags.EmulatorFlagContractAddressFile.GetValue(),
			"deployerKeyFile", chainflags.EmulatorFlagDeployerKeyFile.GetValue(),
			"autoUpdateConnector", chainflags.EmulatorFlagAutoUpdateConnector.GetBool(),
			"config", chainflags.EmulatorFlagConfigFile.GetValue(),
			"contractAddressFile", chainflags.EmulatorFlagContractAddressFile.GetValue(),
		)

		rootCtx := cmd.Context()

		cfg := config.GetGlobalConfig()
		bindings, err := chains.NewChainBindings(rootCtx, cfg, logger)
		if err != nil {
			logger.Error("failed to create chain bindings", "err", err)
			return err
		}
		return bindings.UpdateConnector(rootCtx)
	},
}
