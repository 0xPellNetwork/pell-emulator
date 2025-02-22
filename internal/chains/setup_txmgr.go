package chains

import (
	"github.com/0xPellNetwork/pell-emulator/libs/chains/txmgr"
	"github.com/0xPellNetwork/pell-emulator/libs/chains/wallet"
)

func (cb *ChainBindings) setupDefaultTxMgr() error {
	keyWallet, sender, err := wallet.GetLocalGetWalletByPrivateKey(
		deployerPrivKeyPair,
		cb.RPCClient,
		cb.ChainID,
		cb.logger,
	)
	if err != nil {
		return err
	}

	txMgr := txmgr.NewSimpleTxManager(keyWallet, cb.RPCClient, cb.logger, sender)
	cb.TxMgr = txMgr
	return nil
}
