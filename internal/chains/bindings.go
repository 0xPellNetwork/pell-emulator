package chains

import (
	"github.com/0xPellNetwork/contracts/pkg/contracts/pell_evm/pelldelegationmanager.sol"
	"github.com/0xPellNetwork/contracts/pkg/contracts/pell_evm/pellstrategymanager.sol"
	"github.com/0xPellNetwork/contracts/pkg/contracts/pell_evm/registry/registryrouter.sol"
	"github.com/0xPellNetwork/contracts/pkg/contracts/pell_evm/registry/stakeregistryrouter.sol"
	"github.com/0xPellNetwork/contracts/pkg/contracts/service_evm/omnioperatorsharesmanager.sol"
	"github.com/0xPellNetwork/contracts/pkg/contracts/service_evm/registryinteractor.sol"
	"github.com/0xPellNetwork/contracts/pkg/contracts/staking_evm/core/v2/strategymanager.sol"
	"github.com/0xPellNetwork/contracts/pkg/contracts/staking_evm/core/v3/delegationmanager.sol"
	"github.com/0xPellNetwork/pell-middleware-contracts/pkg/src/centralscheduler.sol"
	"github.com/0xPellNetwork/pell-middleware-contracts/pkg/src/operatorstakemanager.sol"
	gethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"

	"github.com/0xPellNetwork/pell-emulator/config"
	"github.com/0xPellNetwork/pell-emulator/libs/chains/eth"
	"github.com/0xPellNetwork/pell-emulator/libs/log"
)

type typesBinidngs struct {
	// pell evm
	PellDelegationManager   *pelldelegationmanager.PellDelegationManager
	PellRegistryRouter      *registryrouter.RegistryRouter
	PellStrategyManager     *pellstrategymanager.PellStrategyManager
	PellRegistryInteractor  *registryinteractor.RegistryInteractor
	PellStakeRegistryRouter *stakeregistryrouter.StakeRegistryRouter

	// staking evm
	StakingDelegationManager *delegationmanager.DelegationManager
	StakingStrategyManager   *strategymanager.StrategyManager

	// service evm
	ServiceOmniOperatorShareManager *omnioperatorsharesmanager.OmniOperatorSharesManager

	// dvs
	DVSCentralScheduler     *centralscheduler.CentralScheduler
	DVSOperatorStakeManager *operatorstakemanager.OperatorStakeManager
}

type TypesRPCBindings struct {
	typesBinidngs
}

type TypesWsBindings struct {
	typesBinidngs
}

func NewWSBindings(wsClient eth.Client, contractAddress *config.ContractAddress, logger log.Logger) (*TypesWsBindings, error) {
	var wsBds = &TypesWsBindings{}
	var err error

	// pell evm RegistryRouter
	wsBds.PellRegistryRouter, err = registryrouter.NewRegistryRouter(
		gethcommon.HexToAddress(contractAddress.PellRegistryRouter),
		wsClient,
	)
	if err != nil {
		logger.Error("Failed to instantiate a RegistryRouter contract", "error", err)
		return nil, err
	}

	// pell evm PellDelegationManager
	wsBds.PellDelegationManager, err = pelldelegationmanager.NewPellDelegationManager(
		gethcommon.HexToAddress(contractAddress.PellDelegationManager),
		wsClient,
	)
	if err != nil {
		logger.Error("Failed to instantiate a PellDelegationManager contract", "error", err)
		return nil, err
	}

	wsBds.PellStakeRegistryRouter, err = stakeregistryrouter.NewStakeRegistryRouter(
		gethcommon.HexToAddress(contractAddress.PellStakeRegistryRouter),
		wsClient,
	)
	if err != nil {
		logger.Error("Failed to instantiate a StakeRegistryRouter contract", "error", err)
		return nil, errors.Wrap(err, "failed to instantiate a StakeRegistryRouter contract")
	}

	// staking evm OperatorStakeManager
	wsBds.DVSOperatorStakeManager, err = operatorstakemanager.NewOperatorStakeManager(
		gethcommon.HexToAddress(contractAddress.DVSOperatorStakeManager),
		wsClient,
	)
	if err != nil {
		logger.Error("Failed to instantiate a OperatorStakeManager contract", "error", err)
		return nil, err
	}

	// staking evm DelegationManager
	wsBds.StakingDelegationManager, err = delegationmanager.NewDelegationManager(
		gethcommon.HexToAddress(contractAddress.StakingDelegationManager),
		wsClient,
	)
	if err != nil {
		logger.Error("Failed to instantiate a DelegationManager contract", "error", err)
		return nil, err
	}

	if wsBds.StakingDelegationManager == nil {
		return nil, errors.New("Failed to instantiate a DelegationManager contract")
	}

	// staking evm StakingStrategyManager
	wsBds.StakingStrategyManager, err = strategymanager.NewStrategyManager(
		gethcommon.HexToAddress(contractAddress.StakingStrategyManager),
		wsClient,
	)
	if err != nil {
		logger.Error("Failed to instantiate a StakingStrategyManager contract", "error", err)
		return nil, err
	}

	// service evm registryInteractor
	wsBds.PellRegistryInteractor, err = registryinteractor.NewRegistryInteractor(
		gethcommon.HexToAddress(contractAddress.PellRegistryInteractor),
		wsClient,
	)
	if err != nil {
		logger.Error("Failed to instantiate a RegistryInteractor contract", "error", err)
		return nil, err
	}

	// dvs evm DVSCentralScheduler
	wsBds.DVSCentralScheduler, err = centralscheduler.NewCentralScheduler(
		gethcommon.HexToAddress(contractAddress.DVSCentralScheduler),
		wsClient,
	)
	if err != nil {
		logger.Error("Failed to instantiate a CentralScheduler contract", "error", err)
		return nil, err
	}

	logger.Info("ws bindings successfully created", "wsBds", wsBds)

	return wsBds, nil
}

func NewRPCBindings(rpcClient eth.Client, contractAddress *config.ContractAddress, logger log.Logger) (*TypesRPCBindings, error) {
	var rpcBds = &TypesRPCBindings{}
	var err error

	var thisClient = rpcClient

	// pell evm
	// pell evm PellDelegationManager
	rpcBds.PellDelegationManager, err = pelldelegationmanager.NewPellDelegationManager(
		gethcommon.HexToAddress(contractAddress.PellDelegationManager),
		thisClient,
	)
	if err != nil {
		logger.Error("Failed to instantiate a PellDelegationManager contract", "error", err)
		return nil, err
	}

	// pell evm PellRegistryRouter
	rpcBds.PellRegistryRouter, err = registryrouter.NewRegistryRouter(
		gethcommon.HexToAddress(contractAddress.PellRegistryRouter),
		thisClient,
	)
	if err != nil {
		logger.Error("Failed to instantiate a RegistryRouter contract", "error", err)
		return nil, err
	}

	// pell evm PellStrategyManager
	rpcBds.PellStrategyManager, err = pellstrategymanager.NewPellStrategyManager(
		gethcommon.HexToAddress(contractAddress.PellStrategyManager),
		thisClient,
	)
	if err != nil {
		logger.Error("Failed to instantiate a PellStrategyManager contract", "error", err)
		return nil, err
	}

	// staking evm DelegationManager
	delegationManagerAddress := contractAddress.StakingDelegationManager
	rpcBds.StakingDelegationManager, err = delegationmanager.NewDelegationManager(
		gethcommon.HexToAddress(delegationManagerAddress),
		thisClient,
	)
	if err != nil {
		logger.Error("Failed to instantiate a DelegationManager contract", "error", err)
		return nil, err
	}

	// try to update DVS contract address
	//xerr := updateDVSContractAddress(rpcBds)
	//if xerr != nil {
	//	logger.Error("Failed to update DVS contract address from chain, use the one from config", "error", xerr)
	//}

	logger.Info("DVS contract address",
		"CentralScheduler", contractAddress.DVSCentralScheduler,
		"OperatorStakeManager", contractAddress.DVSOperatorStakeManager,
	)

	// service
	rpcBds.ServiceOmniOperatorShareManager, err = omnioperatorsharesmanager.NewOmniOperatorSharesManager(
		gethcommon.HexToAddress(contractAddress.ServiceOmniOperatorSharesManager),
		thisClient,
	)
	if err != nil {
		logger.Error("Failed to instantiate a ServiceManager contract", "error", err)
		return nil, errors.Wrap(err, "failed to instantiate a ServiceManager contract")
	}

	// service evm registryInteractor
	rpcBds.PellRegistryInteractor, err = registryinteractor.NewRegistryInteractor(
		gethcommon.HexToAddress(contractAddress.PellRegistryInteractor),
		thisClient,
	)
	if err != nil {
		logger.Error("Failed to instantiate a RegistryInteractor contract", "error", err)
		return nil, err
	}

	// dvs DVSCentralScheduler
	rpcBds.DVSCentralScheduler, err = centralscheduler.NewCentralScheduler(
		gethcommon.HexToAddress(contractAddress.DVSCentralScheduler),
		thisClient,
	)
	if err != nil {
		logger.Error("Failed to instantiate a CentralScheduler contract", "error", err)
		return nil, err
	}

	// dvs DVSOperatorStakeManager
	rpcBds.DVSOperatorStakeManager, err = operatorstakemanager.NewOperatorStakeManager(
		gethcommon.HexToAddress(contractAddress.DVSOperatorStakeManager),
		thisClient,
	)
	if err != nil {
		logger.Error("Failed to instantiate a OperatorStakeManager contract", "error", err)
		return nil, err
	}

	logger.Info("rpc bindings successfully created", "rpcBds", rpcBds)

	return rpcBds, err
}
