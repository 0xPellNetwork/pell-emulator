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
    sed -i "s|${KEY} = \".*\"|${KEY} = \"${VALUE}\"|" ~/.pelldvs/config/config.toml
  }

  ## update config
  REGISTRY_ROUTER_FACTORY_ADDRESS=$(ssh hardhat "cat $HARDHAT_CONTRACTS_PATH/PellRegistryRouterFactory.json" | jq -r .address)
  PELL_DELEGATION_MNAGER=$(ssh hardhat "cat $HARDHAT_CONTRACTS_PATH/PellDelegationManager-Proxy.json" | jq -r .address)
  PELL_DVS_DIRECTORY=$(ssh hardhat "cat $HARDHAT_CONTRACTS_PATH/PellDVSDirectory-Proxy.json" | jq -r .address)

  update-config rpc_url "$ETH_RPC_URL"
  update-config pell_registry_router_factory_address "$REGISTRY_ROUTER_FACTORY_ADDRESS"
  update-config pell_delegation_manager_address "$PELL_DELEGATION_MNAGER"
  update-config pell_dvs_directory_address "$PELL_DVS_DIRECTORY"
  update-config pell_registry_router_address "$REGISTRY_ROUTER_ADDRESS"
  update-config aggregator_rpc_url "$AGGREGATOR_RPC_URL"
  

  ## FIXME: operator_bls_private_key_store_path should be in the config template. 
  ## FIXME: don't use absolute path for key
  if ! grep -q "operator_bls_private_key_store_path" "$PELLDVS_HOME/config/config.toml"; then
    echo "operator_bls_private_key_store_path = \"$PELLDVS_HOME/keys/$OPERATOR_KEY_NAME.bls.key.json\"" >> $PELLDVS_HOME/config/config.toml
  else
    update-config operator_bls_private_key_store_path "$PELLDVS_HOME/keys/$OPERATOR_KEY_NAME.bls.key.json"
  fi

  ## FIXME: operator_ecdsa_private_key_store_path should be in the config template.
  ## FIXME: don't use absolute path for key
  if ! grep -q "operator_ecdsa_private_key_store_path" "$PELLDVS_HOME/config/config.toml"; then
    echo "operator_ecdsa_private_key_store_path = \"$PELLDVS_HOME/keys/$OPERATOR_KEY_NAME.ecdsa.key.json\"" >> $PELLDVS_HOME/config/config.toml
  else
    update-config operator_ecdsa_private_key_store_path "$PELLDVS_HOME/keys/$OPERATOR_KEY_NAME.ecdsa.key.json"
  fi

#  ## FIXME: why should we use chain.detail.json?
#  scp dvs://$PELLDVS_HOME/config/chain.detail.json $PELLDVS_HOME/config/chain.detail.json
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
