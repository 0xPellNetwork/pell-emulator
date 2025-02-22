package chains

import (
	"context"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/pkg/errors"
)

func (cb *ChainBindings) setupBindings(ctx context.Context) error {
	var err error
	rpcBindings, err := NewRPCBindings(cb.RPCClient, cb.Config.ContractAddress, cb.logger)
	if err != nil {
		cb.logger.Error("Failed to create rpc bindings", "error", err)
		return err
	}
	cb.RPCBindings = rpcBindings

	stakeRegistryRouterAddress, err := rpcBindings.PellRegistryRouter.StakeRegistryRouter(
		&bind.CallOpts{Context: ctx},
	)
	if err != nil {
		cb.logger.Error("failed to get PellStakeRegistryRouterAddress ", "err", err)
		return errors.Wrap(err, "failed to get PellStakeRegistryRouterAddress")
	}
	cb.Config.ContractAddress.PellStakeRegistryRouter = stakeRegistryRouterAddress.String()

	wsBindings, err := NewWSBindings(cb.WsClient, cb.Config.ContractAddress, cb.logger)
	if err != nil {
		cb.logger.Error("Failed to create ws bindings:", "error", err)
		return err
	}

	cb.WsBindings = wsBindings

	return err
}
