package events

import (
	"context"
	"errors"
	"math/big"
	"time"

	"github.com/0xPellNetwork/contracts/pkg/contracts/pell_evm/registry/stakeregistryrouter.sol"
	"github.com/0xPellNetwork/pell-middleware-contracts/pkg/src/operatorstakemanager.sol"
	gethbind "github.com/ethereum/go-ethereum/accounts/abi/bind"

	"github.com/0xPellNetwork/pell-emulator/internal/chains"
	"github.com/0xPellNetwork/pell-emulator/libs/chains/eth"
	"github.com/0xPellNetwork/pell-emulator/libs/chains/txmgr"
	"github.com/0xPellNetwork/pell-emulator/libs/log"
	"github.com/0xPellNetwork/pell-emulator/libs/utils"
)

type EventRegistryRouterSyncAddPools struct {
	BaseEvent
	evtCh chan *stakeregistryrouter.StakeRegistryRouterSyncAddPools
}

func NewEventRegistryRouterSyncAddPools(
	chainID *big.Int,
	rpcClient eth.Client,
	rpcBindings *chains.TypesRPCBindings,
	wsClient eth.Client,
	wsBindings *chains.TypesWsBindings,
	txMgr txmgr.TxManager,
	logger log.Logger,
	hooks ...func(*RegistryInteractorRegisterToPellEvents) error,
) *EventRegistryRouterSyncAddPools {
	eventName := "SyncAddPools"
	contractName := ContractNamePellRegistryRouter
	eventCh := make(chan *stakeregistryrouter.StakeRegistryRouterSyncAddPools)

	var res = &EventRegistryRouterSyncAddPools{
		BaseEvent: BaseEvent{
			srcEVM:       EVMPell,
			eventName:    eventName,
			contractname: contractName,
			chainID:      chainID,
			wsClient:     wsClient,
			rpcClient:    rpcClient,
			wsBindings:   wsBindings,
			rpcBindings:  rpcBindings,
			txMgr:        txMgr,
			targets: []EventTargetInfo{
				newTarget(EVMDVS, "DVSOperatorStakeManager", "SyncAddPools"),
			},
		},
		evtCh: eventCh,
	}

	res.logger = res.setLogger(logger)

	return res
}

func (e *EventRegistryRouterSyncAddPools) process(
	ctx context.Context,
	event *stakeregistryrouter.StakeRegistryRouterSyncAddPools,
) error {
	e.logger.Info("received event: ",
		"groupNumber", event.GroupNumber,
		"poolParams", event.PoolParams,
	)
	groupNumber := event.GroupNumber
	strategyParams := make([]operatorstakemanager.IOperatorStakeManagerPoolParams, len(event.PoolParams))
	for i, strategyParam := range event.PoolParams {
		strategyParams[i] = operatorstakemanager.IOperatorStakeManagerPoolParams(strategyParam)
	}

	e.logger.Info("prepare to forward event ",
		"toContract", "DVSOperatorStakeManager.SyncAddStrategies",
	)

	noSendTxOpts, err := e.txMgr.GetNoSendTxOpts()
	if err != nil {
		return err
	}
	tx, err := e.rpcBindings.DVSOperatorStakeManager.SyncAddPools(noSendTxOpts,
		groupNumber,
		strategyParams,
	)
	if err != nil {
		return err
	}
	receipt, err := e.txMgr.Send(ctx, tx)
	if err != nil {
		return errors.New("failed to send tx with err: " + err.Error())
	}
	e.logger.Info("tx successfully included",
		"txHash", receipt.TxHash.String(),
		"toContract", "DVSOperatorStakeManager.SyncAddStrategies",
	)
	return nil
}

func (e *EventRegistryRouterSyncAddPools) Init(ctx context.Context) error {
	e.logger.Info("init for events")
	sub, err := e.wsBindings.PellStakeRegistryRouter.WatchSyncAddPools(&gethbind.WatchOpts{}, e.evtCh, nil)
	if err != nil {
		e.logger.Error("Failed to subscribe to events", "error", err)
		return err
	}
	e.evtSub = sub
	return nil
}

func (e *EventRegistryRouterSyncAddPools) Listen(ctx context.Context) error {
	e.logger.Info("Listening for events")
	go func(ctx context.Context) {
		for {
			select {
			case event := <-e.evtCh:
				err := e.process(ctx, event)
				if err != nil {
					e.logger.Error("Failed to process to events:", "error", err)
				}
			case err := <-e.evtSub.Err():
				utils.LogSubError(e.logger, err)
				time.Sleep(1 * time.Second)
			case <-ctx.Done():
				e.logger.Info("received unsubscribe signal, shutting down...")
				e.evtSub.Unsubscribe()
				close(e.evtCh)
				return
			default:
				//fmt.Println("Waiting for events...")
				time.Sleep(1 * time.Second)
			}
		}
	}(ctx)
	return nil
}
