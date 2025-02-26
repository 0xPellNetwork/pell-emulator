#!/usr/bin/env bash
set -e

export REGISTRY_ROUTER_ADDRESS_FILE="/root/RegistryRouterAddress.json"

function load_defaults {
  export HARDHAT_CONTRACTS_PATH="/app/pell-middleware-contracts/lib/pell-contracts/deployments/localhost"
  export HARDHAT_CONTRACTS_PATH="/app/pell-middleware-contracts/lib/pell-contracts/deployments/localhost"
  export HARDHAT_PROJ_ROOT="/app/pell-middleware-contracts"
  export HARDHAT_DVS_PATH="/app/pell-middleware-contracts/deployments/localhost"

  export PELLDVS_HOME=${PELLDVS_HOME:-/root/.pelldvs}
  export PELL_EMULATOR_HOME=${PELL_EMULATOR_HOME:-/root/.pell-emulator}
  export ETH_RPC_URL=${ETH_RPC_URL:-http://eth:8545}
  export ETH_WS_URL=${ETH_WS_URL:-ws://eth:8545}
  export ADMIN_KEY_FILE="$PELLDVS_HOME/keys/admin.ecdsa.key.json"
  export REGISTRY_ROUTER_ADDRESS=${REGISTRY_ROUTER_ADDRESS}
}

function create_registry_router {
  ## Create registry router
  # TODO(jimmy): use interactor instead of pelldvs client
  REGISTRY_ROUTER_ADDRESS_FILE="/root/RegistryRouterAddress.json"
  REGISTRY_ROUTER_FACTORY_ADDRESS=$(ssh hardhat "cat $HARDHAT_CONTRACTS_PATH/PellRegistryRouterFactory.json" | jq -r .address)

  # required registry router factory address
  if [ -z "$REGISTRY_ROUTER_FACTORY_ADDRESS" ]; then
    echo "REGISTRY_ROUTER_FACTORY_ADDRESS is required"
    exit 1
  fi

  pelldvs client dvs create-registry-router \
    --home $PELLDVS_HOME \
    --rpc-url $ETH_RPC_URL \
    --from admin \
    --registry-router-factory $REGISTRY_ROUTER_FACTORY_ADDRESS \
    --initial-owner $ADMIN_ADDRESS \
    --dvs-chain-approver $ADMIN_ADDRESS \
    --churn-approver $ADMIN_ADDRESS \
    --ejector $ADMIN_ADDRESS \
    --pauser $ADMIN_ADDRESS \
    --unpauser $ADMIN_ADDRESS \
    --initial-paused-status false \
    --save-to-file $REGISTRY_ROUTER_ADDRESS_FILE \
    --force-save true

  ## Export registry router address
  export REGISTRY_ROUTER_ADDRESS=$(cat $REGISTRY_ROUTER_ADDRESS_FILE | jq -r .address)
}

function setup_registry_router() {
  # if REGISTRY_ROUTER_ADDRESS is set, return
  if [ -n "$REGISTRY_ROUTER_ADDRESS" ]; then
    logt "using existing registry router address from env: $REGISTRY_ROUTER_ADDRESS"
    return
  fi

  # if REGISTRY_ROUTER_ADDRESS_FILE is exits, read it
  if [ -f "$REGISTRY_ROUTER_ADDRESS_FILE" ]; then
    export REGISTRY_ROUTER_ADDRESS=$(cat $REGISTRY_ROUTER_ADDRESS_FILE | jq -r .address)
    logt "using existing registry router address from file: $REGISTRY_ROUTER_ADDRESS"
    return
  fi

  # create
  create_registry_router
}