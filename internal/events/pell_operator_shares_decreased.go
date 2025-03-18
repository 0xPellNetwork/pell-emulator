package events

import (
	"context"
	"math/big"
	"time"

	"github.com/0xPellNetwork/contracts/pkg/contracts/pell_evm/pelldelegationmanager.sol"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	gethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"

	"github.com/0xPellNetwork/pell-emulator/internal/chains"
	"github.com/0xPellNetwork/pell-emulator/libs/chains/eth"
	"github.com/0xPellNetwork/pell-emulator/libs/chains/txmgr"
	"github.com/0xPellNetwork/pell-emulator/libs/log"
	"github.com/0xPellNetwork/pell-emulator/libs/utils"
)

type EventPellDelegationManagerOperatorSharesDecreased struct {
	BaseEvent
	evtChan chan *pelldelegationmanager.PellDelegationManagerOperatorSharesDecreased
}

func NewEventPellDelegationManagerOperatorSharesDecreased(
	chainID *big.Int,
	rpcClient eth.Client,
	rpcBindings *chains.TypesRPCBindings,
	wsClient eth.Client,
	wsBindings *chains.TypesWsBindings,
	txMgr txmgr.TxManager,
	logger log.Logger) *EventPellDelegationManagerOperatorSharesDecreased {

	eventName := "OperatorSharesDecreased"
	contractName := ContractNamePellDelegationManager
	eventCh := make(chan *pelldelegationmanager.PellDelegationManagerOperatorSharesDecreased)

	var res = &EventPellDelegationManagerOperatorSharesDecreased{
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
				newTarget(EVMService, "ServiceOmniOperatorShareManager", "BatchSyncDelegatedShares"),
			},
		},
		evtChan: eventCh,
	}
	res.setLogger(logger)
	return res
}

func (e *EventPellDelegationManagerOperatorSharesDecreased) process(
	ctx context.Context,
	event *pelldelegationmanager.PellDelegationManagerOperatorSharesDecreased,
) error {
	e.logger.Info("received event",
		"ChainId", event.ChainId,
		"Operator", event.Operator,
		"Staker", event.Staker,
		"Pool", event.Strategy,
		"Shares", event.Shares,
	)

	// covert params
	var chainIDs []*big.Int
	var operators []gethcommon.Address
	var pools []gethcommon.Address
	var shares []*big.Int

	chainIDs = append(chainIDs, event.ChainId)
	operators = append(operators, event.Operator)
	pools = append(pools, event.Strategy)
	shares = append(shares, event.Shares)

	noSendTxOpts, err := e.txMgr.GetNoSendTxOpts()
	if err != nil {
		return err
	}
	tx, err := e.rpcBindings.ServiceOmniOperatorShareManager.BatchSyncDelegatedShares(noSendTxOpts,
		chainIDs,
		operators,
		pools,
		shares,
	)
	if err != nil {
		return err
	}
	receipt, err := e.txMgr.Send(ctx, tx)
	if err != nil {
		return errors.New("failed to send tx with err: " + err.Error())
	}
	e.logger.Info("tx successfully included", "txHash", receipt.TxHash.String())

	//updateOperatorsReceipt, err := e.updateOperators(ctx, operators)
	//if err != nil {
	//	return errors.Wrap(err, "failed to update operators")
	//}
	//e.logger.Info("update operators tx successfully included", "txHash", updateOperatorsReceipt.TxHash.String())

	return nil
}

func (e *EventPellDelegationManagerOperatorSharesDecreased) Init(ctx context.Context) error {
	e.logger.Info("init for events")
	sub, err := e.wsBindings.PellDelegationManager.WatchOperatorSharesDecreased(&bind.WatchOpts{},
		e.evtChan, nil, nil,
	)
	if err != nil {
		e.logger.Error("Failed to subscribe to events", "error", err)
		return err
	}

	e.evtSub = sub
	return nil
}

func (e *EventPellDelegationManagerOperatorSharesDecreased) Listen(ctx context.Context) error {
	e.logger.Info("Listening for events")
	go func(ctx context.Context) {
		for {
			select {
			case event := <-e.evtChan:
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
				close(e.evtChan)
				return
			default:
				//fmt.Println("Waiting for events...")
				time.Sleep(1 * time.Second)
			}
		}
	}(ctx)

	return nil
}
