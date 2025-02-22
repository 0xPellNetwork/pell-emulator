package wallet

import (
	"crypto/ecdsa"
	"math/big"
	"os"

	gethcommon "github.com/ethereum/go-ethereum/common"

	"github.com/0xPellNetwork/pell-emulator/libs/chains/eth"
	"github.com/0xPellNetwork/pell-emulator/libs/chains/signerv2"
	"github.com/0xPellNetwork/pell-emulator/libs/log"
)

func GetLocalGetWallet(
	privateKeyStorePath string,
	ethClient eth.Client,
	chainID *big.Int,
	logger log.Logger,
) (Wallet, gethcommon.Address, error) {
	var keyWallet Wallet

	lg := logger.With("module", "wallet/local")

	ecdsaPassword, ok := os.LookupEnv("OPERATOR_ECDSA_KEY_PASSWORD")
	if !ok {
		lg.Info("OPERATOR_ECDSA_KEY_PASSWORD env var not set. using empty string")
	}

	signerCfg := signerv2.Config{
		KeystorePath: privateKeyStorePath,
		Password:     ecdsaPassword,
	}
	sgn, sender, err := signerv2.SignerFromConfig(signerCfg, chainID)
	if err != nil {
		return nil, gethcommon.Address{}, err
	}
	keyWallet, err = NewPrivateKeyWallet(ethClient, sgn, sender, lg)
	if err != nil {
		return nil, gethcommon.Address{}, err
	}
	return keyWallet, sender, nil
}

func GetLocalGetWalletByPrivateKey(
	privateKey *ecdsa.PrivateKey,
	ethClient eth.Client,
	chainID *big.Int, logger log.Logger,
) (Wallet, gethcommon.Address, error) {
	var keyWallet Wallet
	lg := logger.With("module", "wallet/local")

	ecdsaPassword, ok := os.LookupEnv("OPERATOR_ECDSA_KEY_PASSWORD")
	if !ok {
		lg.Info("OPERATOR_ECDSA_KEY_PASSWORD env var not set. using empty string")
	}

	signerCfg := signerv2.Config{
		PrivateKey: privateKey,
		Password:   ecdsaPassword,
	}
	sgn, sender, err := signerv2.SignerFromConfig(signerCfg, chainID)
	if err != nil {
		return nil, gethcommon.Address{}, err
	}

	keyWallet, err = NewPrivateKeyWallet(ethClient, sgn, sender, lg)
	if err != nil {
		return nil, gethcommon.Address{}, err
	}
	return keyWallet, sender, nil
}
