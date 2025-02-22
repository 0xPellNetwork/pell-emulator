package config

import (
	"github.com/0xPellNetwork/pell-emulator/libs/utils"
)

const DefautlHTTPServerPort = 9090

type Config struct {
	Port                int              `json:"port"`
	RPCURL              string           `json:"rpc_url"`
	WSURL               string           `json:"ws_url"`
	ContractAddress     *ContractAddress `json:"contract_address"`
	AutoUpdateConnector bool             `json:"auto_update_connector"`
	DeployerKeyFile     string           `json:"deployer_key_file"`

	LogLevel  string `json:"log_level"`
	LogFormat string `json:"log_format"`
}

func DefaultConfig() *Config {
	return &Config{
		Port:                9090,
		RPCURL:              "http://localhost:8545",
		WSURL:               "ws://localhost:8545",
		ContractAddress:     DefaultContractAddress(),
		AutoUpdateConnector: true,
		LogLevel:            "debug",
		LogFormat:           "plain",
		DeployerKeyFile:     "",
	}
}

func LoadConfigFromFile(filepath string) (*Config, error) {
	var data Config
	err := utils.DecodeJSONFromFile(filepath, &data)
	return &data, err
}
