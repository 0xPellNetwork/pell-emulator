package mocks

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/0xPellNetwork/pell-emulator/cmd/pell-emulator/chainflags"
	"github.com/0xPellNetwork/pell-emulator/config"
	"github.com/0xPellNetwork/pell-emulator/internal/chains"
	"github.com/0xPellNetwork/pell-emulator/libs/log"
)

var logger = log.NewLogger(os.Stdout)

func init() {
	// set flags
	chainflags.FromKeyNameFlag.AddToCmdPersistentFlags(EmulatorMocksCmd)

	// add subcommands
	EmulatorMocksCmd.AddCommand(EmulatorMocksCmdStakingDelegateToCmd)
}

var EmulatorMocksCmd = &cobra.Command{
	Use:   "mocks",
	Short: "pell emulator mocks",
	RunE: func(cmd *cobra.Command, args []string) error {
		_ = cmd.Help()
		return nil
	},
}

func setupForMocksCmd(cmd *cobra.Command) (*chains.ChainBindings, error) {
	cfg := config.GetGlobalConfig()
	bindings, err := chains.NewChainBindings(cmd.Context(), cfg, logger)
	if err != nil {
		return bindings, err
	}

	return bindings, nil
}
