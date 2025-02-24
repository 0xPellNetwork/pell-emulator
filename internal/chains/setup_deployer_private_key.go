package chains

import (
	stdecdsa "crypto/ecdsa"
	"strings"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/pkg/errors"

	"github.com/0xPellNetwork/pell-emulator/libs/chains/crypto/ecdsa"
)

// setupDeployerPrivateKey if keyFilePath is empty, use defaultDeployerPkHex
func (cb *ChainBindings) setupDeployerPrivateKey(keyFilePath string) error {
	if keyFilePath == "" {
		privateKey := strings.TrimPrefix(defaultDeployerPkHex, "0x")
		privateKeyPair, err := crypto.HexToECDSA(privateKey)
		if err != nil {
			return errors.Wrap(err, "failed to decode deployer private key")
		}
		deployerAddress = crypto.PubkeyToAddress(privateKeyPair.PublicKey).Hex()
		deployerPrivKeyPair = privateKeyPair
		return nil
	}

	pk, err := ecdsa.ReadKey(keyFilePath, "")
	if err != nil {
		return errors.Wrap(err, "failed to read deployer private key")
	}

	publicKey := pk.Public()
	publicKeyECDSA, ok := publicKey.(*stdecdsa.PublicKey)
	if !ok {
		return errors.New("error casting public key to ECDSA public key")
	}

	deployerAddress = crypto.PubkeyToAddress(*publicKeyECDSA).Hex()
	//privateKeyHex := hex.EncodeToString(pk.D.Bytes())
	deployerPrivKeyPair = pk

	return nil
}
