package chains

import (
	"context"
	osecdsa "crypto/ecdsa"
	"math/big"

	"github.com/0xPellNetwork/pell-emulator/config"
	"github.com/0xPellNetwork/pell-emulator/libs/chains/eth"
	"github.com/0xPellNetwork/pell-emulator/libs/chains/txmgr"
	"github.com/0xPellNetwork/pell-emulator/libs/log"
)

var (
	deployerAddress      = "0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266"
	defaultDeployerPkHex = "0xac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80"
	deployerPrivKeyPair  *osecdsa.PrivateKey
)

type ChainBindings struct {
	RPCClient   eth.Client
	WsClient    eth.Client
	RPCBindings *TypesRPCBindings
	WsBindings  *TypesWsBindings
	TxMgr       txmgr.TxManager

	ChainID *big.Int

	Config *config.Config

	logger log.Logger
}

func NewChainBindings(ctx context.Context, cfg *config.Config, logger log.Logger) (*ChainBindings, error) {
	var cb = &ChainBindings{
		Config: cfg,
		logger: logger.With("module", "chain-bindings"),
	}
	err := cb.setupClient()
	if err != nil {
		logger.Error("failed to setup client", "error", err)
		return nil, err
	}
	chainID, err := cb.RPCClient.ChainID(ctx)
	if err != nil {
		logger.Error("failed to get chain id", "error", err)
		return nil, err
	}

	cb.ChainID = chainID

	err = cb.setupDeployerPrivateKey(cfg.DeployerKeyFile)
	if err != nil {
		logger.Error("failed to setup deployer private key", "error", err)
		return nil, err
	}

	err = cb.setupDefaultTxMgr()
	if err != nil {
		logger.Error("failed to setup default tx manager", "error", err)
		return nil, err
	}

	err = cb.setupBindings(ctx)
	if err != nil {
		logger.Error("failed to setup bindings", "error", err)
		return nil, err
	}

	if cb.Config.AutoUpdateConnector {
		err = cb.UpdateConnector(ctx)
		if err != nil {
			logger.Error("failed to update connector", "error", err)
			return nil, err
		}
	}

	return cb, nil
}
