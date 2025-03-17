set -e
set -x

function logt() {
  echo "$(date '+%Y-%m-%d %H:%M:%S') $1"
}

function print_new_line {
  echo -e "\n"
}

function assert_eq {
  if [ "$1" != "$2" ]; then
    echo "❌ [FAIL] Expected $1 to be equal to $2"
    exit 1
  fi
  echo "✅ [PASS] Expected $1 to be equal to $2"
}

function asset_gte() {
  local val1=$(printf "%.0f" "$1")
  local val2=$(printf "%.0f" "$2")
  if [ "$val1" -lt "$val2" ]; then
    echo "❌ [FAIL] Expected $1 to be greater than or equal to $2"
    exit 1
  fi
  echo "✅ [PASS] Expected $1 to be greater than or equal to $2"
}

function asset_gt() {
  local val1=$(printf "%.0f" "$1")
  local val2=$(printf "%.0f" "$2")

  if [ "$val1" -le "$val2" ]; then
    echo "❌ [FAIL] Expected $1 to be greater than $2"
    exit 1
  fi
  echo "✅ [PASS] Expected $1 to be greater than $2"
}

function emulator_healthcheck {
  set +e
  set +x
  while true; do
    ready=$(curl -s http://emulator:9090/status | jq -r .ready)
    if [ "$ready" = "true" ]; then
      echo "✅➡️ Emulator initialized, proceeding to the next step..."
      break
    fi
    echo "⌛️ Emulator not initialized, retrying in 2 second..."
    sleep 2
  done
  ## Wait for emulator to be ready
  ## TODO: remove this once we have a better healthcheck
  sleep 8
  set -e
  set -x
}

emulator_healthcheck

# start sshd
/usr/sbin/sshd

if [ ! -f /root/dvs_initialized ]; then
  source "$(dirname "$0")/init_dvs.sh"
  touch /root/dvs_initialized
fi

if [ ! -f /root/operator_initialized ]; then
  source "$(dirname "$0")/init_operator.sh"
  touch /root/operator_initialized
fi

source "$(dirname "$0")/utils.sh"

show_operator_registered $OPERATOR_ADDRESS
print_new_line

show_dvs_operator_info $OPERATOR_ADDRESS
print_new_line

show_operator_stake_status_at_pell_delegation_manager $OPERATOR_ADDRESS
print_new_line

echo "OPERATOR_ADDRESS: " $OPERATOR_ID

logt IS_PELL_OPERATOR
assert_eq "$IS_PELL_OPERATOR" "true"
print_new_line

logt IS_STAKING_OPERATOR
assert_eq "$IS_STAKING_OPERATOR" "true"
print_new_line

logt OPERATOR_STATUS
assert_eq "$OPERATOR_STATUS" "1"
print_new_line

logt operator shares
asset_gt $OPERATOR_STAKE_STATUS 0
print_new_line
