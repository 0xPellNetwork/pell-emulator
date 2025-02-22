package chains

import "github.com/0xPellNetwork/pell-emulator/libs/chains/eth"

func (cb *ChainBindings) setupClient() error {
	var err error
	wsClient, err := eth.NewClient(cb.Config.WSURL)
	if err != nil {
		cb.logger.Error("Failed to connect to the Ethereum wsClient", "error", err)
		return err

	}
	cb.WsClient = wsClient

	rpcClient, err := eth.NewClient(cb.Config.RPCURL)
	if err != nil {
		cb.logger.Error("Failed to connect to the Ethereum rpcClient", "error", err)
		return err
	}

	cb.RPCClient = rpcClient

	return nil
}
