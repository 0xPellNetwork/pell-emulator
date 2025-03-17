package events

import (
	"context"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/0xPellNetwork/contracts/pkg/contracts/pell_evm/registry/registryrouter.sol"
	"github.com/0xPellNetwork/contracts/pkg/contracts/service_evm/registryinteractor.sol"
	gethbind "github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/pkg/errors"

	"github.com/0xPellNetwork/pell-emulator/internal/chains"
	"github.com/0xPellNetwork/pell-emulator/libs/chains/eth"
	"github.com/0xPellNetwork/pell-emulator/libs/chains/txmgr"
	"github.com/0xPellNetwork/pell-emulator/libs/log"
	"github.com/0xPellNetwork/pell-emulator/libs/utils"
)

type EventCentralSchedulerToPell struct {
	BaseEvent
	evtCh                     chan *registryinteractor.RegistryInteractorRegisterCentralSchedulerToPell
	hooksAfterGetAllEventData []func(context.Context, *RegistryInteractorRegisterToPellEvents) error
}

func NewEventCentralSchedulerToPell(
	chainID *big.Int,
	rpcClient eth.Client,
	rpcBindings *chains.TypesRPCBindings,
	wsClient eth.Client,
	wsBindings *chains.TypesWsBindings,
	txMgr txmgr.TxManager,
	logger log.Logger,
	hooks ...func(*RegistryInteractorRegisterToPellEvents) error,
) *EventCentralSchedulerToPell {
	eventName := "CentralSchedulerEvent"
	contractName := "PellRegistryInteractor"

	eventCh := make(chan *registryinteractor.RegistryInteractorRegisterCentralSchedulerToPell)

	var res = &EventCentralSchedulerToPell{
		BaseEvent: BaseEvent{
			srcEVM:       EVMDVS,
			eventName:    eventName,
			contractname: contractName,
			chainID:      chainID,
			wsClient:     wsClient,
			rpcClient:    rpcClient,
			wsBindings:   wsBindings,
			rpcBindings:  rpcBindings,
			txMgr:        txMgr,
			targets: []EventTargetInfo{
				newTarget(EVMPell, "PellRegistryRouter", "AddSupportedChain"),
			},
		},
		evtCh:                     eventCh,
		hooksAfterGetAllEventData: nil,
	}
	res.setLogger(logger)
	return res
}

type RegistryInteractorRegisterToPellEvents struct {
	CentralSchedulerEvent     *registryinteractor.RegistryInteractorRegisterCentralSchedulerToPell
	OperatorStakeManagerEvent *registryinteractor.RegistryInteractorRegisterStakeManagerToPell
	EjectionManagerEvent      *registryinteractor.RegistryInteractorRegisterEjectionManagerToPell
}

func (e *EventCentralSchedulerToPell) Init(ctx context.Context) error {
	e.logger.Info("init for events")
	sub, err := e.wsBindings.PellRegistryInteractor.WatchRegisterCentralSchedulerToPell(&gethbind.WatchOpts{}, e.evtCh)
	if err != nil {
		return err
	}
	e.evtSub = sub
	return nil
}

func (e *EventCentralSchedulerToPell) Listen(ctx context.Context) error {
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
				time.Sleep(1 * time.Second)
			}
		}
	}(ctx)

	return nil
}

func (e *EventCentralSchedulerToPell) process(
	ctx context.Context,
	event *registryinteractor.RegistryInteractorRegisterCentralSchedulerToPell,
) error {
	txHash := event.Raw.TxHash.Hex()
	startBlockNumber := event.Raw.BlockNumber
	endBlockNumber := event.Raw.BlockNumber + 1000

	e.logger.Info("Processing event",
		"txHash", txHash,
		"event.blockNumber", event.Raw.BlockNumber,
		"startBlockNumber", startBlockNumber,
		"endBlockNumber", endBlockNumber,
	)

	allEData, err := e.getAllEventData(txHash, startBlockNumber, endBlockNumber)
	if err != nil {
		e.logger.Error("Failed to get all event data", "error", err)
		return errors.Wrap(err, "failed to get all event data")
	}

	if len(e.hooksAfterGetAllEventData) > 0 {
		for _, hook := range e.hooksAfterGetAllEventData {
			if err := hook(ctx, allEData); err != nil {
				e.logger.Error("hook failed", "error", err)
			}
		}
	}

	dvsInfo := registryrouter.IRegistryRouterDVSInfo{
		ChainId:          e.chainID,
		CentralScheduler: allEData.CentralSchedulerEvent.CentralScheduler,
		EjectionManager:  allEData.EjectionManagerEvent.EjectionManager,
		StakeManager:     allEData.OperatorStakeManagerEvent.StakeManager,
	}

	e.logger.Debug("dvsInfo", "dvsInfo", dvsInfo)

	var dvsChainApproverSignature = registryrouter.ISignatureUtilsSignatureWithSaltAndExpiry{
		Signature: allEData.CentralSchedulerEvent.DvsChainApproverSignature.Signature,
		Salt:      allEData.CentralSchedulerEvent.DvsChainApproverSignature.Salt,
		Expiry:    allEData.CentralSchedulerEvent.DvsChainApproverSignature.Expiry,
	}

	e.logger.Debug("dvsChainApproverSignature", "dvsChainApproverSignature", dvsChainApproverSignature)

	noSendTxOpts, err := e.txMgr.GetNoSendTxOpts()
	if err != nil {
		return errors.Wrap(err, "failed to get no send tx opts")
	}
	tx, err := e.rpcBindings.PellRegistryRouter.AddSupportedChain(noSendTxOpts, dvsInfo, dvsChainApproverSignature)
	if err != nil {
		// if the chain is already supported, we can ignore the error
		if strings.Contains(err.Error(), "revert: RR25") {
			e.logger.Info("chain already supported")
			return nil
		}
		return errors.Wrap(err, "failed to add supported chain")
	}
	receipt, err := e.txMgr.Send(ctx, tx)
	if err != nil {
		return errors.New("failed to send tx with err: " + err.Error())
	}
	e.logger.Info("tx successfully included", "txHash", receipt.TxHash.String())

	return nil
}

func (e *EventCentralSchedulerToPell) getAllEventData(txHash string, startBlock, toBlock uint64) (*RegistryInteractorRegisterToPellEvents, error) {
	// Update the map to use the new struct
	eventMap := make(map[string]*RegistryInteractorRegisterToPellEvents)

	// Filter and process all three types of events
	if err := e.filterCentralSchedulerEvent(startBlock, toBlock, eventMap); err != nil {
		return nil, err
	}

	if err := e.filterOperatorStakeManagerEvent(startBlock, toBlock, eventMap); err != nil {
		return nil, err
	}

	if err := e.filterEjectionManagerEvent(startBlock, toBlock, eventMap); err != nil {
		return nil, err
	}

	for tx, v := range eventMap {
		if tx != txHash {
			e.logger.Info("skipping", "tx", tx)
			continue
		}
		if v.CentralSchedulerEvent != nil && v.OperatorStakeManagerEvent != nil && v.EjectionManagerEvent != nil {
			e.logger.Info("found")
			return v, nil
		}
	}

	return nil, errors.New("failed to get all event data")
}

func (e *EventCentralSchedulerToPell) filterCentralSchedulerEvent(startBlock, toBlock uint64, eventMap map[string]*RegistryInteractorRegisterToPellEvents) error {
	iter, err := e.rpcBindings.PellRegistryInteractor.FilterRegisterCentralSchedulerToPell(&gethbind.FilterOpts{
		Start:   startBlock,
		End:     &toBlock,
		Context: context.TODO(),
	})

	if err != nil {
		e.logger.Error("Failed to filter CentralSchedulerEvent event",
			"error", err,
			"startBlock", startBlock,
			"toBlock", toBlock,
			"chainID", e.chainID,
		)
		return err
	}
	defer iter.Close()

	for iter.Next() {
		event := iter.Event

		txHash := event.Raw.TxHash.Hex()
		e.logger.Info(fmt.Sprintf("CentralSchedulerEvent type event tx hash: %s", txHash))

		if _, exists := eventMap[txHash]; !exists {
			eventMap[txHash] = &RegistryInteractorRegisterToPellEvents{}
		}
		eventMap[txHash].CentralSchedulerEvent = event
	}

	return nil
}

func (e *EventCentralSchedulerToPell) filterOperatorStakeManagerEvent(startBlock, toBlock uint64, eventMap map[string]*RegistryInteractorRegisterToPellEvents) error {
	iter, err := e.rpcBindings.PellRegistryInteractor.FilterRegisterStakeManagerToPell(&gethbind.FilterOpts{
		Start:   startBlock,
		End:     &toBlock,
		Context: context.TODO(),
	})

	if err != nil {
		e.logger.Error("Failed to filter OperatorStakeManagerEvent event",
			"error", err,
			"startBlock", startBlock,
			"toBlock", toBlock,
			"chainID", e.chainID,
		)
		return err
	}
	defer iter.Close()

	for iter.Next() {
		event := iter.Event

		txHash := event.Raw.TxHash.Hex()
		e.logger.Info(fmt.Sprintf("OperatorStakeManagerEvent type event tx hash: %s", txHash))

		if _, exists := eventMap[txHash]; !exists {
			eventMap[txHash] = &RegistryInteractorRegisterToPellEvents{}
		}
		eventMap[txHash].OperatorStakeManagerEvent = event
	}

	return nil
}

func (e *EventCentralSchedulerToPell) filterEjectionManagerEvent(startBlock, toBlock uint64, eventMap map[string]*RegistryInteractorRegisterToPellEvents) error {
	iter, err := e.rpcBindings.PellRegistryInteractor.FilterRegisterEjectionManagerToPell(&gethbind.FilterOpts{
		Start:   startBlock,
		End:     &toBlock,
		Context: context.TODO(),
	})

	if err != nil {
		e.logger.Error("Failed to filter RegisterEjectionManagerToPell event",
			"error", err,
			"startBlock", startBlock,
			"toBlock", toBlock,
			"chainID", e.chainID,
		)
		return err
	}
	defer iter.Close()

	for iter.Next() {
		event := iter.Event

		txHash := event.Raw.TxHash.Hex()
		e.logger.Info(fmt.Sprintf("RegisterEjectionManagerToPell type event tx hash: %s", txHash))

		if _, exists := eventMap[txHash]; !exists {
			eventMap[txHash] = &RegistryInteractorRegisterToPellEvents{}
		}
		eventMap[txHash].EjectionManagerEvent = event
	}

	return nil
}
