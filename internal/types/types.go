package types

import (
	"github.com/0xPellNetwork/pell-emulator/libs/utils"
)

type TypesHardhatDeploymentsContractAddressPell struct {
	DelegationManagerImplementation         string `json:"DelegationManager-Implementation"`
	DelegationManagerProxy                  string `json:"DelegationManager-Proxy"`
	EmptyContract                           string `json:"EmptyContract"`
	GasSwapPEVM                             string `json:"GasSwapPEVM"`
	GatewayPEVM                             string `json:"GatewayPEVM"`
	PauserRegistry                          string `json:"PauserRegistry"`
	PellConnectorOnPell                     string `json:"PellConnectorOnPell"`
	PellDVSDirectoryImplementation          string `json:"PellDVSDirectory-Implementation"`
	PellDVSDirectoryProxy                   string `json:"PellDVSDirectory-Proxy"`
	PellDelegationManagerImplementation     string `json:"PellDelegationManager-Implementation"`
	PellDelegationManagerProxy              string `json:"PellDelegationManager-Proxy"`
	PellEmptyContract                       string `json:"PellEmptyContract"`
	PellProxyAdmin                          string `json:"PellProxyAdmin"`
	PellRegistryRouterImplementation        string `json:"PellRegistryRouter-Implementation"`
	PellRegistryRouterBeacon                string `json:"PellRegistryRouterBeacon"`
	PellRegistryRouterFactory               string `json:"PellRegistryRouterFactory"`
	PellSlasherImplementation               string `json:"PellSlasher-Implementation"`
	PellSlasherProxy                        string `json:"PellSlasher-Proxy"`
	PellStakeRegistryRouterImplementation   string `json:"PellStakeRegistryRouter-Implementation"`
	PellStakeRegistryRouterBeacon           string `json:"PellStakeRegistryRouterBeacon"`
	PellStrategyManagerImplementation       string `json:"PellStrategyManager-Implementation"`
	PellStrategyManagerProxy                string `json:"PellStrategyManager-Proxy"`
	ProxyAdmin                              string `json:"ProxyAdmin"`
	SlasherImplementation                   string `json:"Slasher-Implementation"`
	SlasherProxy                            string `json:"Slasher-Proxy"`
	StrategyImplementation                  string `json:"Strategy-Implementation"`
	StrategyManagerImplementation           string `json:"StrategyManager-Implementation"`
	StrategyManagerProxy                    string `json:"StrategyManager-Proxy"`
	SystemContract                          string `json:"SystemContract"`
	WrappedPell                             string `json:"WrappedPell"`
	PBTCStrategyProxy                       string `json:"pBTC-Strategy-Proxy"`
	PBTCTestnetMintableERC20                string `json:"pBTC-TestnetMintableERC20"`
	StBTCStrategyProxy                      string `json:"stBTC-Strategy-Proxy"`
	StBTCTestnetMintableERC20               string `json:"stBTC-TestnetMintableERC20"`
	TSSManager                              string `json:"TSSManager"`
	PellConnectorOnService                  string `json:"PellConnectorOnService"`
	OmniOperatorSharesManagerImplementation string `json:"OmniOperatorSharesManager-Implementation"`
	OmniOperatorSharesManagerProxy          string `json:"OmniOperatorSharesManager-Proxy"`
	GatewayEVM                              string `json:"GatewayEVM"`
	PellToken                               string `json:"PellToken"`
	UniV2Factory                            string `json:"UniV2Factory"`
	UniV2Pair                               string `json:"UniV2Pair"`
	WrappedToken                            string `json:"WrappedToken"`
	UniV2Router                             string `json:"UniV2Router"`
	GasSwapEVM                              string `json:"GasSwapEVM"`
	RegistryInteractor                      string `json:"RegistryInteractor"`
}

func LoadTypesHardhatDeploymentsContractAddressPellFromFile(filepath string) (*TypesHardhatDeploymentsContractAddressPell, error) {
	var data TypesHardhatDeploymentsContractAddressPell
	err := utils.DecodeJSONFromFile(filepath, &data)
	return &data, err
}

type TypesHardhatDeploymentsContractAddressDVS struct {
	EmptyContract                       string `json:"EmptyContract"`
	ProxyAdmin                          string `json:"ProxyAdmin"`
	MockDVSServiceManagerProxy          string `json:"MockDVSServiceManager-Proxy"`
	CentralSchedulerProxy               string `json:"CentralScheduler-Proxy"`
	OperatorKeyManagerProxy             string `json:"OperatorKeyManager-Proxy"`
	OperatorIndexManagerProxy           string `json:"OperatorIndexManager-Proxy"`
	OperatorStakeManagerProxy           string `json:"OperatorStakeManager-Proxy"`
	OperatorStakeManagerImplementation  string `json:"OperatorStakeManager-Implementation"`
	OperatorKeyManagerImplementation    string `json:"OperatorKeyManager-Implementation"`
	OperatorIndexManagerImplementation  string `json:"OperatorIndexManager-Implementation"`
	CentralSchedulerImplementation      string `json:"CentralScheduler-Implementation"`
	EjectionManagerImplementation       string `json:"EjectionManager-Implementation"`
	EjectionManagerProxy                string `json:"EjectionManager-Proxy"`
	OperatorInfoProvider                string `json:"OperatorInfoProvider"`
	MockDVSServiceManagerImplementation string `json:"MockDVSServiceManager-Implementation"`
}

func LoadTypesHardhatDeploymentsContractAddressDVSFromFile(filepath string) (*TypesHardhatDeploymentsContractAddressDVS, error) {
	var data TypesHardhatDeploymentsContractAddressDVS
	err := utils.DecodeJSONFromFile(filepath, &data)
	return &data, err
}
