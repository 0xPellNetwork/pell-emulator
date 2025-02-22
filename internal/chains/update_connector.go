package chains

import (
	"context"
	"fmt"

	gethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"
)

func (cb *ChainBindings) UpdateConnector(ctx context.Context) error {
	var contractName = "DVSCentralScheduler"
	cb.logger.Info("start update connector for ", "contract", contractName)
	noSendTxOpts, err := cb.TxMgr.GetNoSendTxOpts()
	if err != nil {
		return err
	}
	tx, err := cb.RPCBindings.DVSCentralScheduler.UpdateConnector(
		noSendTxOpts,
		gethcommon.HexToAddress(deployerAddress),
	)
	if err != nil {
		cb.logger.Error("failed to update connector for ", "contract", contractName, "error", err)
		return errors.Wrap(err, fmt.Sprintf("failed to update connector for %s", contractName))
	}
	receipt, err := cb.TxMgr.Send(context.Background(), tx)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("failed to send tx for %s", contractName))
	}

	cb.logger.Info("update connector successfully for ",
		"contract", contractName,
		"txHash", receipt.TxHash.String(),
	)

	/*
		====
	*/
	contractName = "ServiceOmniOperatorShareManager"
	cb.logger.Info("start update connector for ", "contract", contractName)
	noSendTxOpts, err = cb.TxMgr.GetNoSendTxOpts()
	if err != nil {
		return err
	}
	tx, err = cb.RPCBindings.ServiceOmniOperatorShareManager.UpdateConnector(
		noSendTxOpts,
		gethcommon.HexToAddress(deployerAddress),
	)
	if err != nil {
		cb.logger.Error("failed to update connector for ", "contract", contractName, "error", err)
		return errors.Wrap(err, fmt.Sprintf("failed to update connector for %s", contractName))
	}
	receipt, err = cb.TxMgr.Send(context.Background(), tx)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("failed to send tx for %s", contractName))
	}

	cb.logger.Info("update connector successfully for ",
		"contract", contractName,
		"txHash", receipt.TxHash.String(),
	)

	/*
		====
	*/
	contractName = "StakingDelegationManager"
	cb.logger.Info("start update connector for ", "contract", contractName)
	noSendTxOpts, err = cb.TxMgr.GetNoSendTxOpts()
	if err != nil {
		return err
	}
	tx, err = cb.RPCBindings.StakingDelegationManager.UpdateConnector(
		noSendTxOpts,
		gethcommon.HexToAddress(deployerAddress),
	)
	if err != nil {
		cb.logger.Error("failed to update connector for ", "contract", contractName, "error", err)
		return errors.Wrap(err, fmt.Sprintf("failed to update connector for %s", contractName))
	}
	receipt, err = cb.TxMgr.Send(context.Background(), tx)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("failed to send tx for %s", contractName))
	}

	cb.logger.Info("update connector successfully for ",
		"contract", contractName,
		"txHash", receipt.TxHash.String(),
	)

	return nil
}
