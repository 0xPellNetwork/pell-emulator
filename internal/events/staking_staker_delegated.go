package events

import (
	"context"
	"math/big"
	"time"

	"github.com/0xPellNetwork/contracts/pkg/contracts/staking_evm/core/v3/delegationmanager.sol"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	gethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"

	"github.com/0xPellNetwork/pell-emulator/internal/chains"
	"github.com/0xPellNetwork/pell-emulator/libs/chains/eth"
	"github.com/0xPellNetwork/pell-emulator/libs/chains/txmgr"
	"github.com/0xPellNetwork/pell-emulator/libs/log"
	"github.com/0xPellNetwork/pell-emulator/libs/utils"
)

type EventStakingStakerDelegated struct {
	BaseEvent
	evtCh chan *delegationmanager.DelegationManagerStakerDelegated
}

func NewEventStakingStakerDelegated(
	chainID *big.Int,
	rpcClient eth.Client,
	rpcBindings *chains.TypesRPCBindings,
	wsClient eth.Client,
	wsBindings *chains.TypesWsBindings,
	txMgr txmgr.TxManager,
	logger log.Logger,
) *EventStakingStakerDelegated {
	eventName := "StakerDelegated"
	contractName := ContractNameStakingDelegationManager
	eventCh := make(chan *delegationmanager.DelegationManagerStakerDelegated)
	var res = &EventStakingStakerDelegated{
		BaseEvent: BaseEvent{
			srcEVM:      EVMStaking,
			eventName:   eventName,
			srcContract: contractName,
			logger:      logger.With("event", eventName, "contract", contractName),
			chainID:     chainID,
			wsClient:    wsClient,
			rpcClient:   rpcClient,
			wsBindings:  wsBindings,
			rpcBindings: rpcBindings,
			txMgr:       txMgr,
			targets: []EventTargetInfo{
				newTarget(EVMPell, "PellDelegationManager", "SyncDelegateState"),
			},
		},
		evtCh: eventCh,
	}
	res.setLogger(logger)
	return res
}

func (e *EventStakingStakerDelegated) process(
	ctx context.Context, event *delegationmanager.DelegationManagerStakerDelegated,
) error {
	e.logger.Info("received event",
		"Staker", event.Staker,
		"Operator", event.Operator,
	)
	noSendTxOpts, err := e.txMgr.GetNoSendTxOpts()
	if err != nil {
		return err
	}
	tx, err := e.rpcBindings.PellDelegationManager.SyncDelegateState(noSendTxOpts,
		e.chainID,
		event.Staker,
		event.Operator,
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

func (e *EventStakingStakerDelegated) Init(ctx context.Context) error {
	e.logger.Info("init for events")
	var stakerList = make([]gethcommon.Address, 0)
	var operatorList = make([]gethcommon.Address, 0)

	sub, err := e.wsBindings.StakingDelegationManager.WatchStakerDelegated(&bind.WatchOpts{}, e.evtCh, stakerList, operatorList)
	if err != nil {
		e.logger.Error("Failed to subscribe to events", "error", err)
		return err
	}
	e.evtSub = sub
	return nil
}

//nolint:dupl
//nolint:nolintlint
func (e *EventStakingStakerDelegated) Listen(ctx context.Context) error {
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
