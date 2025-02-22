package chains

import (
	"context"
	"crypto/rand"
	"math/big"
	"strings"

	"github.com/0xPellNetwork/contracts/pkg/contracts/pell_evm/registry/registryrouter.sol"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	gethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/pkg/errors"

	"github.com/0xPellNetwork/pell-emulator/libs/log"
)

//nolint:unused
func (cb *ChainBindings) setupChainAddSupportedChain(logger log.Logger) error {
	lg := logger.With("action", "setupChainAddSupportedChain")
	ctx := context.Background()
	var err error

	// generate a random salt and 1 hour expiry for the signature
	var sigSalt [32]byte
	_, err = rand.Read(sigSalt[:])
	if err != nil {
		return errors.Wrap(err, "failed to generate random salt")
	}

	curBlockNum, err := cb.RPCClient.BlockNumber(ctx)
	if err != nil {
		lg.Error("failed to get current block number", "err", err)
		return errors.Wrap(err, "failed to get current block number")
	}
	curBlock, err := cb.RPCClient.BlockByNumber(ctx, big.NewInt(int64(curBlockNum))) //nolint:gosec
	if err != nil {
		lg.Error("failed to get current block", "err", err)
		return errors.Wrap(err, "failed to get current block")
	}

	sigValidForSeconds := int64(60 * 60)                                 // 1 hour
	sigExpiry := big.NewInt(int64(curBlock.Time()) + sigValidForSeconds) //nolint:gosec

	lg.With("debuging", "debuging").Info(
		"debuing...",
		"k", "v...",
		"sigValidForSeconds", sigValidForSeconds,
		"sigExpiry", sigExpiry,
	)

	ejectionManager, err := cb.RPCBindings.DVSCentralScheduler.EjectionManager(&bind.CallOpts{Context: ctx})
	if err != nil {
		lg.Error("failed to get ejection manager", "err", err)
		return errors.Wrap(err, "failed to get ejection manager")
	}
	stakeRegistry, err := cb.RPCBindings.DVSCentralScheduler.OperatorStakeManager(&bind.CallOpts{Context: ctx})
	if err != nil {
		lg.Error("failed to get stake registry", "err", err)
		return errors.Wrap(err, "failed to get stake registry")
	}

	dvsCentralSchedulerAddress := gethcommon.HexToAddress(cb.Config.ContractAddress.DVSCentralScheduler)

	dvsInfos := registryrouter.IRegistryRouterDVSInfo{
		ChainId:          cb.ChainID,
		CentralScheduler: dvsCentralSchedulerAddress,
		EjectionManager:  ejectionManager,
		StakeManager:     stakeRegistry,
	}

	lg.With("debuging", "debuging").Info(
		"debuing...",
		"k", "v...",
		"ejectionManager", ejectionManager,
		"stakeRegistry", stakeRegistry,
		"chainID", cb.ChainID,
		"centralScheduler", dvsCentralSchedulerAddress,
		"dvsInfos", dvsInfos,
		"sigSalt", sigSalt,
		"sigExpiry", sigExpiry,
		"registryRouterAddress", cb.Config.ContractAddress.PellRegistryRouter,
	)

	msgToSign, err := cb.RPCBindings.PellRegistryRouter.CalculateAddSupportedDVSApprovalDigestHash(
		&bind.CallOpts{Context: ctx},
		dvsInfos,
		sigSalt,
		sigExpiry,
	)
	if err != nil {
		lg.Error("failed to calculate add supported dvs approval digest hash", "err", err)
		return errors.Wrap(err, "failed to calculate add supported dvs approval digest hash")
	}

	lg.With("debuging", "debuging").Info(
		"debuing...",
		"k", "v...",
		"msgToSign", msgToSign,
	)

	signature, err := crypto.Sign(msgToSign[:], deployerPrivKeyPair)
	if err != nil {
		return errors.Wrap(err, "failed to sign message")
	}
	// the crypto library is low level and deals with 0/1 v values, whereas ethereum expects 27/28, so we add 27
	// see https://github.com/ethereum/go-ethereum/issues/28757#issuecomment-1874525854
	// and https://twitter.com/pcaversaccio/status/1671488928262529031
	signature[64] += 27

	var dvsChainApproverSignature = registryrouter.ISignatureUtilsSignatureWithSaltAndExpiry{
		Signature: signature,
		Salt:      sigSalt,
		Expiry:    sigExpiry,
	}

	noSendTxOpts, err := cb.TxMgr.GetNoSendTxOpts()
	if err != nil {
		return errors.Wrap(err, "failed to get no send tx opts")
	}
	tx, err := cb.RPCBindings.PellRegistryRouter.AddSupportedChain(noSendTxOpts, dvsInfos, dvsChainApproverSignature)
	if err != nil {
		// if the chain is already supported, we can ignore the error
		if strings.Contains(err.Error(), "revert: RR25") {
			return nil
		}
		return errors.Wrap(err, "failed to add supported chain")
	}
	receipt, err := cb.TxMgr.Send(ctx, tx)
	if err != nil {
		return errors.New("failed to send tx with err: " + err.Error())
	}
	lg.Info("tx successfully included", "txHash", receipt.TxHash.String())

	return nil
}
