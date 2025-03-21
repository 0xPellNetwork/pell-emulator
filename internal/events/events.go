package events

import (
	"context"
	"fmt"
	"math/big"
	"strings"

	gethevent "github.com/ethereum/go-ethereum/event"

	"github.com/0xPellNetwork/pell-emulator/internal/chains"
	"github.com/0xPellNetwork/pell-emulator/libs/chains/eth"
	"github.com/0xPellNetwork/pell-emulator/libs/chains/txmgr"
	"github.com/0xPellNetwork/pell-emulator/libs/log"
)

type IEvents interface {
	Init(ctx context.Context) error
	Listen(ctx context.Context) error
}

type BaseEvent struct {
	srcEVM      string
	eventName   string
	srcContract string
	logger      log.Logger
	chainID     *big.Int
	wsClient    eth.Client
	rpcClient   eth.Client
	wsBindings  *chains.TypesWsBindings
	rpcBindings *chains.TypesRPCBindings
	txMgr       txmgr.TxManager
	evtSub      gethevent.Subscription
	targets     []EventTargetInfo
}

func (be *BaseEvent) setLogger(logger log.Logger) log.Logger {
	targetsInfos := make([]string, 0, len(be.targets))
	for _, target := range be.targets {
		targetsInfos = append(targetsInfos, fmt.Sprintf("target_%s:%s:%s", target.EVM, target.Contract, target.Method))
	}
	be.logger = logger.With(
		"event", be.eventName,
		"srcEvm", be.srcEVM,
		"srcContract", be.srcContract,
		"targets", strings.Join(targetsInfos, ","),
	)

	return be.logger
}

type EventTargetInfo struct {
	EVM      string
	Contract string
	Method   string
}

func newTarget(evm string, contract string, method string) EventTargetInfo {
	return EventTargetInfo{
		EVM:      evm,
		Contract: contract,
		Method:   method,
	}
}

func GetAllEvents(chainID *big.Int,
	rpcClient eth.Client, rpcBindings *chains.TypesRPCBindings,
	wsClient eth.Client, wsBindings *chains.TypesWsBindings,
	txMgr txmgr.TxManager, logger log.Logger) []IEvents {

	var eventList []IEvents

	// pell evm
	//nolint:stylecheck
	var eventRegistryRouterSyncCreateGroup *EventRegistryRouterSyncCreateGroup = NewEventRegistryRouterSyncCreateGroup(
		chainID, rpcClient, rpcBindings, wsClient, wsBindings, txMgr, logger,
	)
	eventList = append(eventList, eventRegistryRouterSyncCreateGroup)

	//nolint:stylecheck
	var eventPellDelegationManagerOperatorRegistered *EventPellDelegationManagerOperatorRegistered = NewEventPellDelegationManagerOperatorRegistered(
		chainID, rpcClient, rpcBindings, wsClient, wsBindings, txMgr, logger,
	)
	eventList = append(eventList, eventPellDelegationManagerOperatorRegistered)

	//nolint:stylecheck
	var eventRegistryRouterSyncRegisterOperator *EventRegistryRouterSyncRegisterOperator = NewEventRegistryRouterSyncRegisterOperator(
		chainID, rpcClient, rpcBindings, wsClient, wsBindings, txMgr, logger,
	)
	eventList = append(eventList, eventRegistryRouterSyncRegisterOperator)

	//nolint:stylecheck
	var eventRegistryRouterSyncUpdateOperators *EventRegistryRouterSyncUpdateOperators = NewEventRegistryRouterSyncUpdateOperators(
		chainID, rpcClient, rpcBindings, wsClient, wsBindings, txMgr, logger,
	)
	eventList = append(eventList, eventRegistryRouterSyncUpdateOperators)

	//nolint:stylecheck
	var eventCentralSchedulerToPell *EventCentralSchedulerToPell = NewEventCentralSchedulerToPell(
		chainID, rpcClient, rpcBindings, wsClient, wsBindings, txMgr, logger,
	)
	eventList = append(eventList, eventCentralSchedulerToPell)

	// staking evm - Deposit
	//nolint:stylecheck
	var eventStakingDeposit *EventStakingDeposit = NewEventStakingDeposit(
		chainID, rpcClient, rpcBindings, wsClient, wsBindings, txMgr, logger,
	)
	eventList = append(eventList, eventStakingDeposit)

	//nolint:stylecheck
	var eventStakingStakerDelegated *EventStakingStakerDelegated = NewEventStakingStakerDelegated(
		chainID, rpcClient, rpcBindings, wsClient, wsBindings, txMgr, logger,
	)
	eventList = append(eventList, eventStakingStakerDelegated)

	//nolint:stylecheck
	var eventStakingStakerUndelegated *EventStakingStakerUndelegated = NewEventStakingStakerUndelegated(
		chainID, rpcClient, rpcBindings, wsClient, wsBindings, txMgr, logger,
	)
	eventList = append(eventList, eventStakingStakerUndelegated)

	//nolint:stylecheck
	var eventStakingWithdrawalQueued *EventStakingWithdrawalQueued = NewEventStakingWithdrawalQueued(
		chainID, rpcClient, rpcBindings, wsClient, wsBindings, txMgr, logger,
	)
	eventList = append(eventList, eventStakingWithdrawalQueued)

	//nolint:stylecheck
	var eventRegistryRouterSyncAddPools *EventRegistryRouterSyncAddPools = NewEventRegistryRouterSyncAddPools(
		chainID, rpcClient, rpcBindings, wsClient, wsBindings, txMgr, logger,
	)
	eventList = append(eventList, eventRegistryRouterSyncAddPools)

	//nolint:stylecheck
	var eventEventPellDelegationManagerOperatorSharesIncreased *EventPellDelegationManagerOperatorSharesIncreased = NewEventPellDelegationManagerOperatorSharesIncreased(
		chainID, rpcClient, rpcBindings, wsClient, wsBindings, txMgr, logger,
	)
	eventList = append(eventList, eventEventPellDelegationManagerOperatorSharesIncreased)

	//nolint:stylecheck
	var eventEventPellDelegationManagerOperatorSharesDecreased *EventPellDelegationManagerOperatorSharesDecreased = NewEventPellDelegationManagerOperatorSharesDecreased(
		chainID, rpcClient, rpcBindings, wsClient, wsBindings, txMgr, logger,
	)
	eventList = append(eventList, eventEventPellDelegationManagerOperatorSharesDecreased)

	return eventList
}
