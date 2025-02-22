# Pell Emulator

Pell Emulator is a Pell DVS event emulator designed to simulate the behavior of the Pell Chain in relation to DVS synchronization operations within a development environment. It listens for specific contract events and forwards them to other contracts.

This emulator fully supports the following key operations:
- Stakers staking ERC-20 tokens to Pell Chain.
- Stakers delegating and undelegating from Pell Chain.
- DVS project builders registering a DVS project and adding a supported chain for their project.
- Registering an operator to Pell Chain and registering an operator to a DVS project.

All event initialization, listening, handling, and forwarding are fully automated. The emulator is designed to provide a complete local development environment for developers without requiring an actual Pell chain to be started. The full list of supported events can be found in the code: [events.go](https://github.com/0xPellNetwork/pell-emulator/blob/main/internal/events/events.go#L34-L105).

## Usage

This guide provides instructions for using the Pell Emulator. It does not include steps for deploying Pell chain contracts or DVS contracts. If you want to run the full process locally, refer to the end-to-end CI workflow here: [e2e.yml](https://github.com/0xPellNetwork/pell-emulator/blob/main/.github/workflows/e2e.yml).

### Clone the Repository

```
git clone https://github.com/0xPellNetwork/pell-emulator && cd pell-emulator
```

### Install Pell Emulator

```
go install cmd/pell-emulator
```

### Initialize Pell Emulator

```
pell-emulator init --home .pell-emulator
```

This command generates a configuration file `config/config.json` in the `.pell-emulator` directory, which looks like this:

```
{
  "port": 9090,
  "rpc_url": "http://localhost:8545",
  "ws_url": "ws://localhost:8545",
  "contract_address": {
    "PellDelegationManager": "0x7a2088a1bFc9d81c55368AE168C2C02570cB814F",
    "PellRegistryRouter": "0x3E69aeCb6a5abAc2D87d6707649E2fB0173ee2Da",
    "PellRegistryInteractor": "0x922D6956C99E12DFeB3224DEA977D0939758A1Fe",
    "StakingStrategyManager": "0xCf7Ed3AccA5a467e9e704C703E8D87F634fB0Fc9",
    "StakingDelegationManager": "0xDc64a140Aa3E981100a9becA4E685f962f0cF6C9",
    "ServiceOmniOperatorSharesManager": "0x4c5859f0F772848b2D91F1D83E2Fe57935348029",
    "DVSCentralScheduler": "0x04C89607413713Ec9775E14b954286519d836FEf",
    "DVSOperatorStakeManager": "0x2E2Ed0Cfd3AD2f1d34481277b3204d807Ca2F8c2"
  },
  "auto_update_connector": true,
  "deployer_key_file": "",
  "log_level": "debug",
  "log_format": "plain"
}
```

### Build Docker Image

```
GITHUB_TOKEN=<YOUR_GITHUB_TOKEN> make docker-build
```

### Start Docker Containers

```
HOST_PELL_EMULATOR_HOME=.pell-emulator make docker-up-all
```

### Stop Docker Containers

```
make docker-down
```

### Modify Configuration

Edit the configuration file `.pell-emulator/config/config.json` to specify the RPC endpoint of the Pell blockchain, contract addresses, and other settings.

### Update Connector

This operation updates the connector contract address in the service chain:

```
pell-emulator update-connector --home .pell-emulator
```

### Start Pell Emulator

```
pell-emulator start --home .pell-emulator  
```

## Development

To contribute to Pell Emulator, clone the repository:

```
git clone https://github.com/0xPellNetwork/pell-emulator
```

For contribution guidelines, please refer to the [CONTRIBUTING.md](./CONTRIBUTING.md) document.
