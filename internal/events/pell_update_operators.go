package events

import (
	"context"
	"math/big"
	"time"

	"github.com/0xPellNetwork/contracts/pkg/contracts/pell_evm/registry/registryrouter.sol"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	gethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"

	"github.com/0xPellNetwork/pell-emulator/internal/chains"
	"github.com/0xPellNetwork/pell-emulator/libs/chains/eth"
	"github.com/0xPellNetwork/pell-emulator/libs/chains/txmgr"
	"github.com/0xPellNetwork/pell-emulator/libs/log"
	"github.com/0xPellNetwork/pell-emulator/libs/utils"
)

type EventRegistryRouterSyncUpdateOperators struct {
	BaseEvent
	evtCh chan *registryrouter.RegistryRouterSyncUpdateOperators
}

func NewEventRegistryRouterSyncUpdateOperators(
	chainID *big.Int,
	rpcClient eth.Client,
	rpcBindings *chains.TypesRPCBindings,
	wsClient eth.Client,
	wsBindings *chains.TypesWsBindings,
	txMgr txmgr.TxManager,
	logger log.Logger,
) *EventRegistryRouterSyncUpdateOperators {

	eventName := "SyncUpdateOperators"
	contractName := ContractNamePellRegistryRouter

	eventCh := make(chan *registryrouter.RegistryRouterSyncUpdateOperators)
	var res = &EventRegistryRouterSyncUpdateOperators{
		BaseEvent: BaseEvent{
			srcEVM:       EVMPell,
			eventName:    eventName,
			contractname: contractName,
			logger:       logger.With("event", eventName, "contract", contractName),
			chainID:      chainID,
			wsClient:     wsClient,
			rpcClient:    rpcClient,
			wsBindings:   wsBindings,
			rpcBindings:  rpcBindings,
			txMgr:        txMgr,
			targets: []EventTargetInfo{
				newTarget(EVMDVS, "DVSCentralScheduler", "SyncUpdateOperators"),
			},
		},
		evtCh: eventCh,
	}
	res.setLogger(logger)
	return res
}

func (e *EventRegistryRouterSyncUpdateOperators) process(
	ctx context.Context, event *registryrouter.RegistryRouterSyncUpdateOperators,
) error {
	e.logger.Info("received event",
		"Operators", event.Operators,
	)

	if e.rpcBindings.DVSCentralScheduler == nil {
		return errors.New("DVSCentralScheduler is nil")
	}

	// covert params
	operators := make([]gethcommon.Address, len(event.Operators))

	//nolint:gosimple
	for i, operator := range event.Operators {
		operators[i] = operator
	}

	noSendTxOpts, err := e.txMgr.GetNoSendTxOpts()
	if err != nil {
		return err
	}
	tx, err := e.rpcBindings.DVSCentralScheduler.SyncUpdateOperators(noSendTxOpts,
		operators,
	)
	if err != nil {
		return err
	}
	receipt, err := e.txMgr.Send(ctx, tx)
	if err != nil {
		return errors.Wrap(err, "failed to send tx")
	}

	e.logger.Info("tx successfully included", "txHash", receipt.TxHash.String())

	return nil
}

func (e *EventRegistryRouterSyncUpdateOperators) Init(ctx context.Context) error {
	e.logger.Info("init for events")
	sub, err := e.wsBindings.PellRegistryRouter.WatchSyncUpdateOperators(&bind.WatchOpts{}, e.evtCh)
	if err != nil {
		e.logger.Error("Failed to subscribe to events", "error", err)
		return err
	}
	e.evtSub = sub
	return nil
}

func (e *EventRegistryRouterSyncUpdateOperators) Listen(ctx context.Context) error {
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
