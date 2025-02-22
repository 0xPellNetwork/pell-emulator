package mocks

import (
	"github.com/spf13/cobra"

	"github.com/0xPellNetwork/pell-emulator/cmd/pell-emulator/chainflags"
	"github.com/0xPellNetwork/pell-emulator/config"
	"github.com/0xPellNetwork/pell-emulator/libs/chains/crypto/ecdsa"
	"github.com/0xPellNetwork/pell-emulator/libs/utils"
)

func init() {

	chainflags.EmulatorFlagOperatorAddress.AddToCmdFlag(EmulatorMocksCmdStakingDelegateToCmd)

	KeyFileFlag.AddToCmdFlag(EmulatorMocksCmdStakingDelegateToCmd)

	// mark required flags
	_ = chainflags.MarkFlagsAreRequired(EmulatorMocksCmdStakingDelegateToCmd,
		chainflags.FromKeyNameFlag,
		chainflags.EmulatorFlagOperatorAddress,
	)
}

var EmulatorMocksCmdStakingDelegateToCmd = &cobra.Command{
	Use:   "staking-delegate-to",
	Short: "pell_emulator mocks staking delegate to",
	Example: `
	pelldvs pell_emulator mocks staking-delegate-to --from ba01 --operator 0x70997970C51812dc3A010C7d01b50e0d17dc79C8
`,
	RunE: func(cmd *cobra.Command, args []string) error {

		cfg := config.GetGlobalConfig()
		var bindings, err = setupForMocksCmd(cmd)
		if err != nil {
			return err
		}

		logger.Info("mocks staking delegate to",
			"k", "v",
			"keyName", chainflags.FromKeyNameFlag.Value,
			"operator", chainflags.EmulatorFlagOperatorAddress.Value,
			"rpcURL", cfg.RPCURL,
		)

		pk, err := ecdsa.ReadKey(KeyFileFlag.Value, "")
		if err != nil {
			return err
		}

		myTxMgr, err := utils.CreateTxMgrByKeyFile(pk, bindings.RPCClient, bindings.ChainID, logger)
		if err != nil {
			return err
		}
		mgr, err := NewStakingDelegationManager(cfg.RPCURL, bindings.RPCClient, pk, myTxMgr, cfg.ContractAddress.StakingDelegationManager)
		if err != nil {
			return err
		}

		receipt, err := mgr.DelegateTo(cmd.Context(), chainflags.EmulatorFlagOperatorAddress.Value)
		if err != nil {
			return err
		}

		if receipt == nil && err == nil {
			return nil
		}

		logger.Info("delegated to operator", "receipt", receipt.TxHash.String())

		_ = cmd.Help()
		return nil
	},
}
