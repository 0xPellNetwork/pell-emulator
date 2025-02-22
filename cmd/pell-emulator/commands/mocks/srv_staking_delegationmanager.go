package mocks

import (
	"context"
	"crypto/ecdsa"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"math/big"
	"os"
	"time"

	"github.com/0xPellNetwork/contracts/pkg/contracts/staking_evm/core/v3/delegationmanager.sol"
	gethcommon "github.com/ethereum/go-ethereum/common"
	gethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/pkg/errors"

	"github.com/0xPellNetwork/pell-emulator/libs/chains/eth"
	"github.com/0xPellNetwork/pell-emulator/libs/chains/txmgr"
	"github.com/0xPellNetwork/pell-emulator/libs/utils/exec"
)

type StakingDelegationManager struct {
	contract   *delegationmanager.DelegationManager
	rpcClient  eth.Client
	txMgr      txmgr.TxManager
	rpcURL     string
	address    string
	privateKey *ecdsa.PrivateKey
}

func NewStakingDelegationManager(rpcURL string, rpcClient eth.Client, privateKey *ecdsa.PrivateKey, txMgr txmgr.TxManager, address string) (*StakingDelegationManager, error) {
	sdm := &StakingDelegationManager{

		rpcClient:  rpcClient,
		txMgr:      txMgr,
		rpcURL:     rpcURL,
		address:    address,
		privateKey: privateKey,
	}
	var err error
	sdm.contract, err = delegationmanager.NewDelegationManager(gethcommon.HexToAddress(address), rpcClient)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create delegation manager")
	}
	return sdm, nil
}

func (sdm *StakingDelegationManager) DelegateTo(ctx context.Context, operator string) (*gethtypes.Receipt, error) {

	lg := logger.With("comps", "MocksStaking.DelegationManager")
	lg.Info("Delegating to operator", "operator", operator)

	// prepare params
	// generate a valid salt and expiry
	var sigSalt [32]byte
	_, err := rand.Read(sigSalt[:])
	if err != nil {
		lg.Error("failed to generate random salt", "err", err)
		return nil, errors.Wrap(err, "failed to generate random salt")
	}

	sigValidForSeconds := int64(60 * 60) // 1 hour
	sigExpiry := big.NewInt(time.Now().Unix() + sigValidForSeconds)
	approverSignatureAndExpiry := delegationmanager.ISignatureUtilsSignatureWithExpiry{
		Signature: nil,
		Expiry:    sigExpiry,
	}

	lg.Info("params",
		"operator", operator,
		"sigSalt", sigSalt,
		"sigExpiry", sigExpiry,
		"approverSignatureAndExpiry", approverSignatureAndExpiry,
	)

	funcSignature := "delegateTo(address,(bytes,uint256),bytes32)"
	args := []string{
		"cast", "send", // command
		"--rpc-url", sdm.rpcURL, // rpc url
		sdm.address,   // contract address
		funcSignature, // function signature
		operator,      // operator address
		"(0x0000000000000000000000000000000000000000000000000000000000000000,0)", // approverSignatureAndExpiry
		fmt.Sprintf("0x%x", sigSalt),                                  // sigSalt
		"--private-key", hex.EncodeToString(sdm.privateKey.D.Bytes()), // private key
	}

	logger.Info("executing command", "args", args)

	err = exec.CommandVerbose(ctx, args...)

	if os.Getenv("USEGO") == "" {
		return nil, err
	}

	noSendTxOpts, err := sdm.txMgr.GetNoSendTxOpts()
	if err != nil {
		lg.Error("failed to get no send tx opts", "err", err)
		return nil, errors.Wrap(err, "failed to get no send tx opts")
	}

	tx, err := sdm.contract.DelegateTo(noSendTxOpts,
		gethcommon.HexToAddress(operator),
		approverSignatureAndExpiry,
		sigSalt,
	)

	if err != nil {
		lg.Error("failed to create tx", "err", err)
		return nil, errors.Wrap(err, "failed to create tx")
	}

	receipt, err := sdm.txMgr.Send(ctx, tx)
	if err != nil {
		lg.Error("failed to send tx", "err", err)
		return nil, errors.Wrap(err, "failed to send tx")
	}

	lg.Info("tx successfully included", "txHash", receipt.TxHash.String())

	return receipt, nil
}
