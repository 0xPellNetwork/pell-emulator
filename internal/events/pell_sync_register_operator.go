package events

import (
	"context"
	"errors"
	"math/big"
	"time"

	"github.com/0xPellNetwork/contracts/pkg/contracts/pell_evm/registry/registryrouter.sol"
	"github.com/0xPellNetwork/pell-middleware-contracts/pkg/src/centralscheduler.sol"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	gethcommon "github.com/ethereum/go-ethereum/common"

	"github.com/0xPellNetwork/pell-emulator/internal/chains"
	"github.com/0xPellNetwork/pell-emulator/libs/chains/eth"
	"github.com/0xPellNetwork/pell-emulator/libs/chains/txmgr"
	"github.com/0xPellNetwork/pell-emulator/libs/log"
	"github.com/0xPellNetwork/pell-emulator/libs/utils"
)

type EventRegistryRouterSyncRegisterOperator struct {
	BaseEvent
	evtCh chan *registryrouter.RegistryRouterSyncRegisterOperator
}

func NewEventRegistryRouterSyncRegisterOperator(
	chainID *big.Int,
	rpcClient eth.Client,
	rpcBindings *chains.TypesRPCBindings,
	wsClient eth.Client,
	wsBindings *chains.TypesWsBindings,
	txMgr txmgr.TxManager,
	logger log.Logger,
) *EventRegistryRouterSyncRegisterOperator {

	eventName := "SyncRegisterOperator"
	contractName := ContractNamePellRegistryRouter
	eventCh := make(chan *registryrouter.RegistryRouterSyncRegisterOperator)

	var res = &EventRegistryRouterSyncRegisterOperator{
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

func (e *EventRegistryRouterSyncRegisterOperator) process(
	ctx context.Context, event *registryrouter.RegistryRouterSyncRegisterOperator,
) error {
	e.logger.Info("received event",
		"Operator", event.Operator,
		"OperatorID", event.OperatorId,
		"GroupNumbers", event.GroupNumbers,
		"Socket", event.Socket,
		"PubKeyParams", event.Params,
	)

	if e.rpcBindings.DVSCentralScheduler == nil {
		return errors.New("DVSCentralScheduler is nil")
	}

	// covert params from event
	var operatorAddress gethcommon.Address
	var groupNumbers []byte
	var pubKeyParams centralscheduler.IOperatorKeyManagerPubkeyRegistrationParams

	operatorAddress = event.Operator
	groupNumbers = event.GroupNumbers
	pubKeyParams = centralscheduler.IOperatorKeyManagerPubkeyRegistrationParams{
		PubkeyG1: centralscheduler.BN254G1Point{
			X: event.Params.PubkeyG1.X,
			Y: event.Params.PubkeyG1.Y,
		},
		PubkeyG2: centralscheduler.BN254G2Point{
			X: event.Params.PubkeyG2.X,
			Y: event.Params.PubkeyG2.Y,
		},
	}

	noSendTxOpts, err := e.txMgr.GetNoSendTxOpts()
	if err != nil {
		return err
	}
	tx, err := e.rpcBindings.DVSCentralScheduler.SyncRegisterOperator(noSendTxOpts,
		operatorAddress,
		groupNumbers,
		pubKeyParams,
	)
	if err != nil {
		return err
	}
	receipt, err := e.txMgr.Send(ctx, tx)
	if err != nil {
		return errors.New("failed to send tx with err: " + err.Error())
	}
	e.logger.Info("tx successfully included", "txHash", receipt.TxHash.String())

	return nil
}

func (e *EventRegistryRouterSyncRegisterOperator) Init(ctx context.Context) error {
	e.logger.Info("init for events")
	sub, err := e.wsBindings.PellRegistryRouter.WatchSyncRegisterOperator(
		&bind.WatchOpts{},
		e.evtCh,
		nil,
		nil,
	)
	if err != nil {
		e.logger.Error("Failed to subscribe to events", "error", err)
		return err
	}

	e.evtSub = sub

	return nil
}

func (e *EventRegistryRouterSyncRegisterOperator) Listen(ctx context.Context) error {
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
