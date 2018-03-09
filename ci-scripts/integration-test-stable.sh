#!/bin/bash
# Runs "stable"-mode tests against a mdl node configured with a pinned database
# "stable" mode tests assume the blockchain data is static, in order to check API responses more precisely
# $TEST defines which test to run i.e, cli or gui; If empty both are run

#Set Script Name variable
SCRIPT=`basename ${BASH_SOURCE[0]}`
PORT="46420"
RPC_PORT="46430"
IP_ADDR="0.0.0.0"
HOST="http://$IP_ADDR:$PORT"
RPC_ADDR="$IP_ADDR:$RPC_PORT"
MODE="stable"
BINARY="$PWD/mdl-integration"
TEST=""
UPDATE=""

usage () {
  echo "Usage: $SCRIPT"
  echo "Optional command line arguments"
  echo "-t <string>  -- Test to run, gui or cli; empty runs both tests"
  echo "-u <boolean> -- Update stable testdata"
  exit 1
}

while getopts "h?t:u" args; do
case $args in
    h|\?)
        usage;
        exit;;
    t ) TEST=${OPTARG};;
    u ) UPDATE="--update";;
  esac
done

set -euxo pipefail

DATA_DIR=$(mktemp -d -t mdl-data-dir.XXXXXX)
WALLET_DIR="${DATA_DIR}/wallets"

if [[ ! "$DATA_DIR" ]]; then
  echo "Could not create temp dir"
  exit 1
fi

# Compile the mdl node
# We can't use "go run" because this creates two processes which doesn't allow us to kill it at the end
echo "compiling mdl"
go build -o "$BINARY" $PWD/cmd/mdl/mdl.go

# Run mdl node with pinned blockchain database
echo "starting mdl ($PWD/mdl-integration) node in background with http listener on $HOST"

$PWD/mdl-integration -disable-networking=true \
                      -web-interface-addr=$IP_ADDR \
                      -web-interface-port=$PORT \
                      -download-peerlist=false \
                      -db-path=./src/gui/integration/test-fixtures/blockchain-development.db \
                      -db-read-only=true \
                      -rpc-interface=true \
                      -rpc-interface-addr=$IP_ADDR \
                      -rpc-interface-port=$RPC_PORT \
                      -launch-browser=false \
                      -data-dir="$DATA_DIR" \
                      -wallet-dir="$WALLET_DIR" &
MDL_PID=$!

echo "mdl node pid=$MDL_PID"

echo "sleeping for startup"
sleep 5
echo "done sleeping"

set +e

if [[ -z $TEST || $TEST = "gui" ]]; then

MDL_INTEGRATION_TESTS=1 MDL_INTEGRATION_TEST_MODE=$MODE MDL_NODE_HOST=$HOST go test ./src/gui/integration/... $UPDATE -timeout=60s -v

GUI_FAIL=$?

fi

if [[ -z $TEST  || $TEST = "cli" ]]; then

MDL_INTEGRATION_TESTS=1 MDL_INTEGRATION_TEST_MODE=$MODE RPC_ADDR=$RPC_ADDR go test ./src/api/cli/integration/... $UPDATE -timeout=60s -v

CLI_FAIL=$?

fi


echo "shutting down mdl node"

# Shutdown mdl node
kill -9 $MDL_PID
wait $MDL_PID

rm "$BINARY"


if [[ (-z $TEST || $TEST = "gui") && $GUI_FAIL -ne 0 ]]; then
  exit $GUI_FAIL
elif [[ (-z $TEST || $TEST = "cli") && $CLI_FAIL -ne 0 ]]; then
  exit $CLI_FAIL
else 
  exit 0
fi
# exit $FAIL
