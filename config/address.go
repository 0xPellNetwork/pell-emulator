package config

import (
	"github.com/0xPellNetwork/pell-emulator/libs/utils"
)

type ContractAddress struct {
	// pell evm
	PellDelegationManager   string `json:"PellDelegationManager"`
	PellRegistryRouter      string `json:"PellRegistryRouter"`
	PellRegistryInteractor  string `json:"PellRegistryInteractor"`
	PellStakeRegistryRouter string `json:"-"`

	// staking evm
	StakingStrategyManager   string `json:"StakingStrategyManager"`
	StakingDelegationManager string `json:"StakingDelegationManager"`

	// service evm
	ServiceOmniOperatorSharesManager string `json:"ServiceOmniOperatorSharesManager"`

	// dvs
	DVSCentralScheduler     string `json:"DVSCentralScheduler"`
	DVSOperatorStakeManager string `json:"DVSOperatorStakeManager"`
}

var DefaultContractAddress = func() *ContractAddress {
	return &ContractAddress{
		// pell evm
		PellDelegationManager:  "0x7a2088a1bFc9d81c55368AE168C2C02570cB814F",
		PellRegistryRouter:     "0x3E69aeCb6a5abAc2D87d6707649E2fB0173ee2Da",
		PellRegistryInteractor: "0x922D6956C99E12DFeB3224DEA977D0939758A1Fe",

		// staking evm
		StakingStrategyManager:   "0xCf7Ed3AccA5a467e9e704C703E8D87F634fB0Fc9",
		StakingDelegationManager: "0xDc64a140Aa3E981100a9becA4E685f962f0cF6C9",

		// service evm
		ServiceOmniOperatorSharesManager: "0x4c5859f0F772848b2D91F1D83E2Fe57935348029",

		// dvs
		DVSCentralScheduler:     "0x04C89607413713Ec9775E14b954286519d836FEf",
		DVSOperatorStakeManager: "0x2E2Ed0Cfd3AD2f1d34481277b3204d807Ca2F8c2",
	}
}

func LoadContractAddressFromFile(filepath string) (*ContractAddress, error) {
	var data ContractAddress
	err := utils.DecodeJSONFromFile(filepath, &data)
	return &data, err
}
