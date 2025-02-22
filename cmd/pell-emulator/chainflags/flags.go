package chainflags

var FromKeyNameFlag = &StringFlag{
	Name:    "from",
	Usage:   "Name of the key to use for signing the transaction",
	EnvVars: []string{"PELL_FROM"},
}

var EthRPCURLFlag = &StringFlag{
	Name:    "rpc-url",
	Usage:   "URL of the Ethereum RPC server",
	EnvVars: []string{"PELL_ETH_RPC_URL"},
}

var PellRegistryRouterFactoryAddress = &StringFlag{
	Name:    "registry-router-factory",
	Usage:   "ContractAddress of the registry router factory contract",
	EnvVars: []string{"PELL_PELL_REGISTRY_ROUTER_FACTORY"},
}

var PellDelegationManagerAddress = &StringFlag{
	Name:    "delegation-manager",
	Usage:   "ContractAddress of the delegation contract",
	EnvVars: []string{"PELL_PELL_DELEGATION"},
}

var PellRegistryRouterAddress = &StringFlag{
	Name:    "registry-router",
	Usage:   "ContractAddress of the registry router contract",
	EnvVars: []string{"PELL_PELL_REGISTRY_ROUTER"},
}

var PellDvsDirectoryAddress = &StringFlag{
	Name:    "dvs-directory",
	Usage:   "ContractAddress of the DVS directory contract",
	EnvVars: []string{"PELL_PELL_DVS_DIRECTORY"},
}

var CentralSchedulerContractAddressFlag = &StringFlag{
	Name:    "central-scheduler",
	Usage:   "ContractAddress of the registry CentralScheduler contract",
	EnvVars: []string{"PELL_CENTRAL_SCHEDULER"},
}

var ChainIDFlag = &IntFlag{
	Name:  "chain-id",
	Usage: "Chain ID",
}

var LogLevelFlag = &StringFlag{
	Name:    "log-level",
	Usage:   "Log level",
	Default: "debug",
}

var LogFormatFlag = &StringFlag{
	Name:    "log-format",
	Usage:   "Log format",
	Default: "plain",
}

// declare flags for PellEmulatorCmd
// rpc url
var EmulatorFlagRPCURL = &StringFlag{
	Name:  "rpc-url",
	Usage: "URL of the RPC server",
}

// wsURL
var EmulatorFlagWSURL = &StringFlag{
	Name:  "ws-url",
	Usage: "URL of the WebSocket server",
}

var EmulatorFlagOperatorAddress = &StringFlag{
	Name:  "operator",
	Usage: "operator address",
}

var EmulatorFlagConfigFile = &StringFlag{
	Name:  "config",
	Usage: "config",
}

var EmulatorFlagContractAddressFile = &StringFlag{
	Name:  "contract-address-file",
	Usage: "contract address file",
}

var EmulatorFlagDeployerKeyFile = &StringFlag{
	Name:  "deployer-key-file",
	Usage: "deployer key file",
}

var EmulatorFlagAutoUpdateConnector = &StringFlag{
	Name:  "auto-update-connector",
	Usage: "auto update connector",
}
