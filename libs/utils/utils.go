package utils

import (
	"crypto/ecdsa"
	"encoding/json"
	"math/big"
	"os"
	"time"

	"github.com/0xPellNetwork/pell-emulator/libs/chains/eth"
	"github.com/0xPellNetwork/pell-emulator/libs/chains/txmgr"
	"github.com/0xPellNetwork/pell-emulator/libs/chains/wallet"
	"github.com/0xPellNetwork/pell-emulator/libs/log"
)

func DecodeJSONFromFile(filepath string, data any) error {
	input, err := os.ReadFile(filepath)
	if err != nil {
		return err
	}
	err = json.Unmarshal(input, data)
	return err
}

//nolint:unused
//nolint:nolintlint
func WriteJSONToFile(filepath string, data any) error {
	output, err := json.Marshal(data)
	if err != nil {
		return err
	}
	err = os.WriteFile(filepath, output, 0644)
	return err
}

//nolint:unused
func writeJSONToFileWithIndent(filepath string, data any) error {
	output, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	err = os.WriteFile(filepath, output, 0644)
	return err
}

func CreateTxMgrByKeyFile(
	privateKey *ecdsa.PrivateKey,
	rpcClient eth.Client,
	chainID *big.Int,
	logger log.Logger,
) (txmgr.TxManager, error) {
	keyWallet, sender, err := wallet.GetLocalGetWalletByPrivateKey(
		privateKey,
		rpcClient,
		chainID,
		logger,
	)
	if err != nil {
		return nil, err
	}
	return txmgr.NewSimpleTxManager(keyWallet, rpcClient, logger, sender), nil
}

func LogSubError(lg log.Logger, err error) {
	if time.Now().Second()%20 == 0 { //nolint: staticcheck
		//lg.Error("Subscription error", "error", err)
	}
}
