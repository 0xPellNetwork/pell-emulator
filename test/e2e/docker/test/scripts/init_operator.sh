logt() {
  echo "$(date '+%Y-%m-%d %H:%M:%S') $1"
}

function load_defaults {
  export OPERATOR_KEY_NAME=${OPERATOR_KEY_NAME:-operator}
  export AGGREGATOR_RPC_URL=${AGGREGATOR_RPC_URL:-dvs:26653}
  export OPERATOR_NODE_NAME=${OPERATOR_NODE_NAME:-operator}
  export REGISTRY_ROUTER_ADDRESS=${REGISTRY_ROUTER_ADDRESS}

  export PELLDVS_HOME=${PELLDVS_HOME:-/root/.pelldvs}
  export ETH_RPC_URL=${ETH_RPC_URL:-http://eth:8545}
  export ETH_WS_URL=${ETH_WS_URL:-ws://eth:8545}

  export AGGREGATOR_INDEXER_START_HEIGHT=${AGGREGATOR_INDEXER_START_HEIGHT:-0}
  export AGGREGATOR_INDEXER_BATCH_SIZE=${AGGREGATOR_INDEXER_BATCH_SIZE:-1000}

  export CHAIN_ID=${CHAIN_ID:-1337}
  export SERVICE_CHAIN_ID=${SERVICE_CHAIN_ID:-1337}
  export SERVICE_CHAIN_RPC_URL=${SERVICE_CHAIN_RPC_URL:-http://eth:8545}

  source "$(dirname "$0")/utils.sh"
}

function set_registry_router_address() {
  # if $REGISTRY_ROUTER_ADDRESS is not set, fetch it from RegistryRouterAddress.json
  if [ -z "$REGISTRY_ROUTER_ADDRESS" ]; then
    export REGISTRY_ROUTER_ADDRESS=$(ssh emulator "cat /root/RegistryRouterAddress.json" | jq -r .address)
  else
    echo "Using provided REGISTRY_ROUTER_ADDRESS: $REGISTRY_ROUTER_ADDRESS"
  fi
}

## TODO: move operator config to seperated location
function init_pelldvs_config {
  pelldvs init --home $PELLDVS_HOME
  update-config() {
    KEY="$1"
    VALUE="$2"
    sed -i "s|${KEY} = \".*\"|${KEY} = \"${VALUE}\"|" $PELLDVS_HOME/config/config.toml
  }

  ## update config
  REGISTRY_ROUTER_FACTORY_ADDRESS=$(ssh hardhat "cat $HARDHAT_CONTRACTS_PATH/PellRegistryRouterFactory.json" | jq -r .address)
  PELL_DELEGATION_MNAGER=$(ssh hardhat "cat $HARDHAT_CONTRACTS_PATH/PellDelegationManager-Proxy.json" | jq -r .address)
  PELL_DVS_DIRECTORY=$(ssh hardhat "cat $HARDHAT_CONTRACTS_PATH/PellDVSDirectory-Proxy.json" | jq -r .address)
  REGISTRY_ROUTER_ADDRESS=$(ssh emulator "cat /root/RegistryRouterAddress.json" | jq -r .address)

  update-config rpc_url "$ETH_RPC_URL"
  update-config aggregator_rpc_url "$AGGREGATOR_RPC_URL"

  ## FIXME: operator_bls_private_key_store_path should be in the config template.
  ## FIXME: don't use absolute path for key
  update-config operator_bls_private_key_store_path "$PELLDVS_HOME/keys/$OPERATOR_KEY_NAME.bls.key.json"
  update-config operator_ecdsa_private_key_store_path "$PELLDVS_HOME/keys/$OPERATOR_KEY_NAME.ecdsa.key.json"

  update-config interactor_config_path "$PELLDVS_HOME/config/interactor_config.json"

  DVS_OPERATOR_KEY_MANAGER=$(ssh hardhat "cat $HARDHAT_DVS_PATH/OperatorKeyManager-Proxy.json" | jq -r .address)
  DVS_CENTRAL_SCHEDULER=$(ssh hardhat "cat $HARDHAT_DVS_PATH/CentralScheduler-Proxy.json" | jq -r .address)
  DVS_OPERATOR_INFO_PROVIDER=$(ssh hardhat "cat $HARDHAT_DVS_PATH/OperatorInfoProvider.json" | jq -r .address)
  DVS_OPERATOR_INDEX_MANAGER=$(ssh hardhat "cat $HARDHAT_DVS_PATH/OperatorIndexManager-Proxy.json" | jq -r .address)

  cat <<EOF > $PELLDVS_HOME/config/interactor_config.json
{
    "rpc_url": "$ETH_RPC_URL",
    "chain_id": $CHAIN_ID,
    "indexer_start_height": $AGGREGATOR_INDEXER_START_HEIGHT,
    "indexer_batch_size": $AGGREGATOR_INDEXER_BATCH_SIZE,
    "contract_config": {
      "pell_registry_router_factory": "$REGISTRY_ROUTER_FACTORY_ADDRESS",
    	"pell_dvs_directory": "$PELL_DVS_DIRECTORY",
    	"pell_delegation_manager": "$PELL_DELEGATION_MNAGER",
    	"pell_registry_router": "$REGISTRY_ROUTER_ADDRESS",
      "dvs_configs": {
        "$CHAIN_ID": {
          "chain_id": $SERVICE_CHAIN_ID,
          "rpc_url": "$SERVICE_CHAIN_RPC_URL",
          "operator_info_provider": "$DVS_OPERATOR_INFO_PROVIDER",
          "operator_key_manager": "$DVS_OPERATOR_KEY_MANAGER",
          "central_scheduler": "$DVS_CENTRAL_SCHEDULER",
          "operator_index_manager": "$DVS_OPERATOR_INDEX_MANAGER"
        }
      }
    }
}
EOF

cat $PELLDVS_HOME/config/interactor_config.json
}


function setup_operator_key {
  if pelldvs keys show $OPERATOR_KEY_NAME --home "$PELLDVS_HOME" >/dev/null 2>&1; then
    echo "Operator key already exists, skipping import"
    return
  fi

  ## Create operator key
  echo  -ne '\n\n' | pelldvs keys create $OPERATOR_KEY_NAME --key-type=ecdsa --insecure > /tmp/operator.key
  export OPERATOR_KEY=$(cat /tmp/operator.key | sed -n 's/.*\/\/[[:space:]]*\([0-9a-f]\{64\}\)[[:space:]]*\/\/.*/\1/p')
  export OPERATOR_ADDRESS=$(pelldvs keys show $OPERATOR_KEY_NAME --home $PELLDVS_HOME | awk '/Key content:/{getline; print}' | head -n 1 | jq -r .address)

  ## To register operator in the DVS, we need the operator's BLS key with the same name
  echo  -ne '\n\n' | pelldvs keys create $OPERATOR_KEY_NAME --key-type=bls --insecure > /tmp/operator_bls.key

  ## Faucet operator with 1 PELL
  faucet $OPERATOR_ADDRESS 1e18
}

function register_operator {
  OPERATOR_METADATA_URI=https://raw.githubusercontent.com/matthew7251/Metadata/main/Matthew_Metadata.json
  pelldvs client operator register-operator \
    --home $PELLDVS_HOME \
    --from $OPERATOR_KEY_NAME \
    --metadata-uri $OPERATOR_METADATA_URI

  show_operator_registered "$OPERATOR_ADDRESS"
}

function register_operator_to_dvs {
  pelldvs client operator register-operator-to-dvs \
    --home $PELLDVS_HOME \
    --from $OPERATOR_KEY_NAME \
    --groups 0 \
    --socket http://$OPERATOR_NODE_NAME:26657
  show_dvs_operator_info $OPERATOR_ADDRESS
}

function stake_and_delegate_to_operator() {
  TestDeployerPrivKey="ac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80"
  TestDeployerEvmAddr="0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266"

  STBTC_ERC20_ADDRESS=$(ssh hardhat "cat $HARDHAT_CONTRACTS_PATH/stBTC-TestnetMintableERC20.json" | jq -r .address)
  STRTEGY_MANAGER_ADDRESS=$(ssh hardhat "cat $HARDHAT_CONTRACTS_PATH/StrategyManager-Proxy.json" | jq -r .address)
  STBTC_STRATEGY_ADDRESS=$(ssh hardhat "cat $HARDHAT_CONTRACTS_PATH/stBTC-Strategy-Proxy.json" | jq -r .address)

  STAKE_AMOUNT=3000000000000000000

  cast call $STBTC_ERC20_ADDRESS "balanceOf(address)(uint256)" $TestDeployerEvmAddr --rpc-url $ETH_RPC_URL

  cast send $STBTC_ERC20_ADDRESS \
    "approve(address,uint256)" \
    $STRTEGY_MANAGER_ADDRESS \
    $STAKE_AMOUNT \
    --private-key $TestDeployerPrivKey \
    --rpc-url $ETH_RPC_URL

  cast send $STRTEGY_MANAGER_ADDRESS \
    "depositIntoStrategy(address,address,uint256)" \
    $STBTC_STRATEGY_ADDRESS \
    $STBTC_ERC20_ADDRESS \
    $STAKE_AMOUNT \
    --private-key $TestDeployerPrivKey \
    --rpc-url $ETH_RPC_URL

  logt "Query Strategy Shares"

  cast call $STRTEGY_MANAGER_ADDRESS \
    "stakerStrategyShares(address,address)(uint256)" \
    $TestDeployerEvmAddr \
    $STBTC_STRATEGY_ADDRESS \
    --rpc-url $ETH_RPC_URL

  logt "Stake to Strategy Done"

  sleep 5

  DELEGATION_MNAGER=$(ssh hardhat "cat $HARDHAT_CONTRACTS_PATH/DelegationManager-Proxy.json" | jq -r .address)
  PELL_DELEGATION_MNAGER=$(ssh hardhat "cat $HARDHAT_CONTRACTS_PATH/PellDelegationManager-Proxy.json" | jq -r .address)
  OPERATOR_ADDRESS=$(pelldvs keys show $OPERATOR_KEY_NAME --home $PELLDVS_HOME | awk '/Key content:/{getline; print}' | head -n 1 | jq -r .address)

  cast send $DELEGATION_MNAGER \
    "delegateTo(address,(bytes,uint256),bytes32)" \
    $OPERATOR_ADDRESS \
    "(0x, 0)" \
    0x0000000000000000000000000000000000000000000000000000000000000000 \
    --private-key $TestDeployerPrivKey \
    --rpc-url $ETH_RPC_URL

  echo -e "\n\n"
  logt "Query Operator Shares"

  sleep 10

  cast call $DELEGATION_MNAGER \
    "getOperatorShares(address,address[])(uint256[])" \
    $OPERATOR_ADDRESS \
    "[$STBTC_STRATEGY_ADDRESS]" \
    --rpc-url $ETH_RPC_URL

  cast call $PELL_DELEGATION_MNAGER \
    "getOperatorShares(address,(uint256,address)[])(uint256[])" \
    $OPERATOR_ADDRESS \
    "[(1337,$STBTC_STRATEGY_ADDRESS)]" \
    --rpc-url $ETH_RPC_URL

  echo -e "\n\n"
  logt "Query group mini stake"

  REGISTRY_ROUTER_ADDRESS=$(ssh emulator "cat /root/RegistryRouterAddress.json" | jq -r .address)
  STAKE_REGISTRY_ROUTER=$(cast call $REGISTRY_ROUTER_ADDRESS "stakeRegistryRouter()(address)" --rpc-url $ETH_RPC_URL)

  cast call $STAKE_REGISTRY_ROUTER \
    "minimumStakeForGroup(uint8)(uint96)" \
    0 \
    --rpc-url $ETH_RPC_URL

  echo -e "\n\n"
  logt "Delegate to Operator Done"
}


logt "Load Default Values for ENV Vars if not set."
load_defaults

logt "Set Registry Router Address"
set_registry_router_address

logt "Init Pelldvs Config from init_operator"
init_pelldvs_config

logt "Setup Operator Key"
setup_operator_key

logt "Register Operator"
register_operator

logt "Register Operator to DVS"
register_operator_to_dvs

logt "stake_and_delegate_to_operator"
stake_and_delegate_to_operator

echo -e "\n\n"
logt "init operator done"
