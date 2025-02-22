package events

import (
	"context"
	"math/big"
	"time"

	"github.com/0xPellNetwork/contracts/pkg/contracts/staking_evm/core/v2/strategymanager.sol"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/pkg/errors"

	"github.com/0xPellNetwork/pell-emulator/internal/chains"
	"github.com/0xPellNetwork/pell-emulator/libs/chains/eth"
	"github.com/0xPellNetwork/pell-emulator/libs/chains/txmgr"
	"github.com/0xPellNetwork/pell-emulator/libs/log"
	"github.com/0xPellNetwork/pell-emulator/libs/utils"
)

type EventStakingDeposit struct {
	BaseEvent
	evtCh chan *strategymanager.StrategyManagerDeposit
}

func NewEventStakingDeposit(
	chainID *big.Int,
	rpcClient eth.Client,
	rpcBindings *chains.TypesRPCBindings,
	wsClient eth.Client,
	wsBindings *chains.TypesWsBindings,
	txMgr txmgr.TxManager,
	logger log.Logger,
) *EventStakingDeposit {

	eventName := "Deposit"
	contractName := "StakingStrategyManager"

	eventCh := make(chan *strategymanager.StrategyManagerDeposit)

	var res = &EventStakingDeposit{
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

func (e *EventStakingDeposit) process(
	ctx context.Context, event *strategymanager.StrategyManagerDeposit,
) error {

	e.logger.Info("received event",
		"Staker", event.Staker,
		"Token", event.Token,
		"Pool", event.Strategy,
		"Shares", event.Shares,
	)

	noSendTxOpts, err := e.txMgr.GetNoSendTxOpts()
	if err != nil {
		return err
	}
	tx, err := e.rpcBindings.PellStrategyManager.SyncDepositState(noSendTxOpts,
		e.chainID,
		event.Staker,
		event.Strategy,
		event.Shares,
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

func (e *EventStakingDeposit) Init(ctx context.Context) error {
	e.logger.Info("init for events")
	sub, err := e.wsBindings.StakingStrategyManager.WatchDeposit(&bind.WatchOpts{}, e.evtCh)
	if err != nil {
		e.logger.Error("Failed to subscribe to events", "error", err)
		return err
	}
	e.evtSub = sub
	return nil
}

func (e *EventStakingDeposit) Listen(ctx context.Context) error {
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
