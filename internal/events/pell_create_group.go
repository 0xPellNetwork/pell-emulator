package events

import (
	"context"
	"errors"
	"math/big"
	"time"

	"github.com/0xPellNetwork/contracts/pkg/contracts/pell_evm/registry/registryrouter.sol"
	"github.com/0xPellNetwork/pell-middleware-contracts/pkg/src/centralscheduler.sol"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"

	"github.com/0xPellNetwork/pell-emulator/internal/chains"
	"github.com/0xPellNetwork/pell-emulator/libs/chains/eth"
	"github.com/0xPellNetwork/pell-emulator/libs/chains/txmgr"
	"github.com/0xPellNetwork/pell-emulator/libs/log"
	"github.com/0xPellNetwork/pell-emulator/libs/utils"
)

type EventRegistryRouterSyncCreateGroup struct {
	BaseEvent
	evtCh chan *registryrouter.RegistryRouterSyncCreateGroup
}

func NewEventRegistryRouterSyncCreateGroup(
	chainID *big.Int,
	rpcClient eth.Client,
	rpcBindings *chains.TypesRPCBindings,
	wsClient eth.Client,
	wsBindings *chains.TypesWsBindings,
	txMgr txmgr.TxManager,
	logger log.Logger) *EventRegistryRouterSyncCreateGroup {

	eventName := "SyncCreateGroup"
	contractName := ContractNamePellRegistryRouter
	eventCh := make(chan *registryrouter.RegistryRouterSyncCreateGroup)

	var res = &EventRegistryRouterSyncCreateGroup{
		BaseEvent: BaseEvent{
			EventName:    eventName,
			Contractname: contractName,
			logger:       logger.With("event", eventName, "contract", contractName),
			chainID:      chainID,
			wsClient:     wsClient,
			rpcClient:    rpcClient,
			wsBindings:   wsBindings,
			rpcBindings:  rpcBindings,
			txMgr:        txMgr,
		},
		evtCh: eventCh,
	}
	return res
}

func (e *EventRegistryRouterSyncCreateGroup) process(
	ctx context.Context, event *registryrouter.RegistryRouterSyncCreateGroup,
) error {

	e.logger.Info("received event",
		"GroupNumber", event.GroupNumber,
		"OperatorSetParams", event.OperatorSetParams,
		"MinimumStake", event.MinimumStake,
	)

	if e.rpcBindings.DVSCentralScheduler == nil {
		return errors.New("DVSCentralScheduler is nil")
	}

	e.logger.Info("prepare to forward event ")
	// covert params
	operatorSetParams := centralscheduler.ICentralSchedulerOperatorSetParam(event.OperatorSetParams)
	mintStake := event.MinimumStake
	poolParams := make([]centralscheduler.IOperatorStakeManagerPoolParams, len(event.PoolParams))
	for i, strategyParam := range event.PoolParams {
		poolParams[i] = centralscheduler.IOperatorStakeManagerPoolParams(strategyParam)
	}

	noSendTxOpts, err := e.txMgr.GetNoSendTxOpts()
	if err != nil {
		return err
	}
	tx, err := e.rpcBindings.DVSCentralScheduler.SyncCreateGroup(noSendTxOpts,
		event.GroupNumber,
		operatorSetParams,
		mintStake,
		poolParams,
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
	)

	return nil
}

func (e *EventRegistryRouterSyncCreateGroup) Init(ctx context.Context) error {
	sub, err := e.wsBindings.PellRegistryRouter.WatchSyncCreateGroup(&bind.WatchOpts{}, e.evtCh, nil)
	if err != nil {
		e.logger.Error("Failed to subscribe to events", "error", err)
		return err
	}
	e.evtSub = sub
	return nil
}

func (e *EventRegistryRouterSyncCreateGroup) Listen(ctx context.Context) error {
	e.logger.Info("Listening for events")

	go func(ctx context.Context) {
		for {
			select {
			case event := <-e.evtCh:
				e.logger.Info("here, received event")
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
