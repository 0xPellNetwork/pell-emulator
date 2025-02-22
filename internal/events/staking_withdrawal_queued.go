package events

import (
	"context"
	"math/big"
	"time"

	"github.com/0xPellNetwork/contracts/pkg/contracts/pell_evm/pelldelegationmanager.sol"
	"github.com/0xPellNetwork/contracts/pkg/contracts/staking_evm/core/v3/delegationmanager.sol"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/pkg/errors"

	"github.com/0xPellNetwork/pell-emulator/internal/chains"
	"github.com/0xPellNetwork/pell-emulator/libs/chains/eth"
	"github.com/0xPellNetwork/pell-emulator/libs/chains/txmgr"
	"github.com/0xPellNetwork/pell-emulator/libs/log"
	"github.com/0xPellNetwork/pell-emulator/libs/utils"
)

type EventStakingWithdrawalQueued struct {
	BaseEvent
	evtCh chan *delegationmanager.DelegationManagerWithdrawalQueued
}

func NewEventStakingWithdrawalQueued(
	chainID *big.Int,
	rpcClient eth.Client,
	rpcBindings *chains.TypesRPCBindings,
	wsClient eth.Client,
	wsBindings *chains.TypesWsBindings,
	txMgr txmgr.TxManager,
	logger log.Logger) *EventStakingWithdrawalQueued {

	eventName := "StakingWithdrawalQueued"
	contractName := ContractNameStakingDelegationManager
	eventCh := make(chan *delegationmanager.DelegationManagerWithdrawalQueued)

	var res = &EventStakingWithdrawalQueued{
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

func (e *EventStakingWithdrawalQueued) process(
	ctx context.Context, event *delegationmanager.DelegationManagerWithdrawalQueued,
) error {

	e.logger.Info("received event",
		"WithdrawalRoot", event.WithdrawalRoot,
		"Withdrawal", event.Withdrawal,
	)

	// covert params from event
	var staker = event.Withdrawal.Staker
	var operator = event.Withdrawal.DelegatedTo
	var withdrawalParams = pelldelegationmanager.IPellDelegationManagerWithdrawalParams{
		Strategies: event.Withdrawal.Strategies,
		Shares:     event.Withdrawal.Shares,
	}

	noSendTxOpts, err := e.txMgr.GetNoSendTxOpts()
	if err != nil {
		return err
	}
	tx, err := e.rpcBindings.PellDelegationManager.SyncWithdrawalState(noSendTxOpts,
		e.chainID,
		staker,
		operator,
		withdrawalParams,
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

func (e *EventStakingWithdrawalQueued) Init(ctx context.Context) error {
	e.logger.Info("init for events")
	eventCh := make(chan *delegationmanager.DelegationManagerWithdrawalQueued)

	sub, err := e.wsBindings.StakingDelegationManager.WatchWithdrawalQueued(&bind.WatchOpts{}, eventCh)
	if err != nil {
		e.logger.Error("Failed to subscribe to events", "error", err)
		return err
	}

	e.evtSub = sub

	return nil
}

func (e *EventStakingWithdrawalQueued) Listen(ctx context.Context) error {
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
