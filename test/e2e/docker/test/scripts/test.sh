set -e

logt() {
  echo "$(date '+%Y-%m-%d %H:%M:%S') $1"
}

function assert_eq {
  if [ "$1" != "$2" ]; then
    echo "❌ [FAIL] Expected $1 to be equal to $2"
    exit 1
  fi
  echo "✅ [PASS] Expected $1 to be equal to $2"
}

function emulator_healthcheck {
  set +e
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

show_dvs_operator_info $OPERATOR_ADDRESS

echo  -e "\n\n"

echo "OPERATOR_ID: " $OPERATOR_ID

assert_eq "$IS_PELL_OPERATOR" "true"
assert_eq "$IS_STAKING_OPERATOR" "true"

assert_eq "$OPERATOR_STATUS" "1"

echo -e "\n\n"