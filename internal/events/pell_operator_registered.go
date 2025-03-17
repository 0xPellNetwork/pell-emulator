package events

import (
	"context"
	"errors"
	"math/big"
	"time"

	"github.com/0xPellNetwork/contracts/pkg/contracts/pell_evm/pelldelegationmanager.sol"
	"github.com/0xPellNetwork/contracts/pkg/contracts/staking_evm/core/v3/delegationmanager.sol"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	gethcommon "github.com/ethereum/go-ethereum/common"

	"github.com/0xPellNetwork/pell-emulator/internal/chains"
	"github.com/0xPellNetwork/pell-emulator/libs/chains/eth"
	"github.com/0xPellNetwork/pell-emulator/libs/chains/txmgr"
	"github.com/0xPellNetwork/pell-emulator/libs/log"
	"github.com/0xPellNetwork/pell-emulator/libs/utils"
)

type EventPellDelegationManagerOperatorRegistered struct {
	BaseEvent
	evtCh chan *pelldelegationmanager.PellDelegationManagerOperatorRegistered
}

func NewEventPellDelegationManagerOperatorRegistered(
	chainID *big.Int,
	rpcClient eth.Client,
	rpcBindings *chains.TypesRPCBindings,
	wsClient eth.Client,
	wsBindings *chains.TypesWsBindings,
	txMgr txmgr.TxManager,
	logger log.Logger) *EventPellDelegationManagerOperatorRegistered {

	eventName := "OperatorRegistered"
	contractName := ContractNamePellDelegationManager

	eventCh := make(chan *pelldelegationmanager.PellDelegationManagerOperatorRegistered)

	var res = &EventPellDelegationManagerOperatorRegistered{
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
				newTarget(EVMDVS, "StakingDelegationManager", "SyncRegisterAsOperator"),
			},
		},
		evtCh: eventCh,
	}
	res.setLogger(logger)
	return res
}

func (e *EventPellDelegationManagerOperatorRegistered) process(
	ctx context.Context,
	event *pelldelegationmanager.PellDelegationManagerOperatorRegistered,
) error {
	e.logger.Info("received event",
		"Operator", event.Operator,
		"Details", event.OperatorDetails,
	)

	// covert params
	operator := event.Operator
	details := delegationmanager.IDelegationManagerOperatorDetails{
		DeprecatedEarningsReceiver: gethcommon.Address{},
		DelegationApprover:         event.OperatorDetails.DelegationApprover,
		StakerOptOutWindow:         event.OperatorDetails.StakerOptOutWindow,
	}

	// TODO(jimmy): DeprecatedEarningsReceiver is not in the event, so it is set to the zero address.

	noSendTxOpts, err := e.txMgr.GetNoSendTxOpts()
	if err != nil {
		return err
	}
	tx, err := e.rpcBindings.StakingDelegationManager.SyncRegisterAsOperator(noSendTxOpts,
		operator,
		details,
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

func (e *EventPellDelegationManagerOperatorRegistered) Init(ctx context.Context) error {
	e.logger.Info("init for events")
	operatorAddressList := make([]gethcommon.Address, 0)
	sub, err := e.wsBindings.PellDelegationManager.WatchOperatorRegistered(&bind.WatchOpts{}, e.evtCh, operatorAddressList)
	if err != nil {
		e.logger.Error("Failed to subscribe to events", "error", err)
		return err
	}

	e.evtSub = sub
	return nil
}

func (e *EventPellDelegationManagerOperatorRegistered) Listen(ctx context.Context) error {
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
