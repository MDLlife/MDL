#!/bin/bash
# Runs "enable-seed-api"-mode tests against a mdl node configured with -enable-seed-api=true
# and /api/v1/wallet/seed api endpoint should return coresponding seed

# Set Script Name variable
SCRIPT=`basename ${BASH_SOURCE[0]}`

# Find unused port
PORT="1024"
while $(lsof -Pi :$PORT -sTCP:LISTEN -t >/dev/null) ; do
    PORT=$((PORT+1))
done

COIN=${COIN:-mdl}
RPC_PORT="$PORT"
HOST="http://127.0.0.1:$PORT"
RPC_ADDR="http://127.0.0.1:$RPC_PORT"
MODE="enable-seed-api"
BINARY="${COIN}-integration-enable-seed-api.test"
TEST=""
RUN_TESTS=""
# run go test with -v flag
VERBOSE=""

usage () {
  echo "Usage: $SCRIPT"
  echo "Optional command line arguments"
  echo "-t <string>  -- Test to run, api or cli; empty runs both tests"
  echo "-v <boolean> -- Run test with -v flag"
  exit 1
}

while getopts "h?t:r:v" args; do
  case $args in
    h|\?)
        usage;
        exit;;
    t ) TEST=${OPTARG};;
    v ) VERBOSE="-v";;
    r ) RUN_TESTS="-run ${OPTARG}";;
  esac
done

COVERAGEFILE="coverage/${BINARY}.coverage.out"
if [ -f "${COVERAGEFILE}" ]; then
    rm "${COVERAGEFILE}"
fi

set -euxo pipefail

CMDPKG=$(go list ./cmd/${COIN})
COVERPKG=$(dirname $(dirname ${CMDPKG}))

DATA_DIR=$(mktemp -d -t ${COIN}-data-dir.XXXXXX)
WALLET_DIR="${DATA_DIR}/wallets"

if [[ ! "$DATA_DIR" ]]; then
  echo "Could not create temp dir"
  exit 1
fi

# Compile the mdl node
# We can't use "go run" because that creates two processes which doesn't allow us to kill it at the end
echo "compiling $COIN with coverage"
go test -c -tags testrunmain -o "$BINARY" -coverpkg="${COVERPKG}/..." ./cmd/${COIN}/

mkdir -p coverage/

# Run mdl node with pinned blockchain database
echo "starting $COIN node in background with http listener on $HOST"

./"$BINARY" -disable-networking=true \
            -genesis-signature eb10468d10054d15f2b6f8946cd46797779aa20a7617ceb4be884189f219bc9a164e56a5b9f7bec392a804ff3740210348d73db77a37adb542a8e08d429ac92700 \
            -genesis-address 2jBbGxZRGoQG1mqhPBnXnLTxK6oxsTf8os6 \
            -blockchain-public-key 0328c576d3f420e7682058a981173a4b374c7cc5ff55bf394d3cf57059bbe6456a \
            -web-interface-port=$PORT \
            -download-peerlist=false \
            -db-path=./src/api/integration/testdata/blockchain-180.db \
            -db-read-only=true \
            -launch-browser=false \
            -data-dir="$DATA_DIR" \
            -wallet-dir="$WALLET_DIR" \
            -enable-all-api-sets=true \
            -enable-api-sets=DEPRECATED_WALLET_SPEND,INSECURE_WALLET_SEED \
            -test.run "^TestRunMain$" \
            -test.coverprofile="${COVERAGEFILE}" \
            &

MDL_PID=$!

echo "$COIN node pid=$MDL_PID"

echo "sleeping for startup"
sleep 3
echo "done sleeping"

set +e

if [[ -z $TEST || $TEST = "api" ]]; then

MDL_INTEGRATION_TESTS=1 MDL_INTEGRATION_TEST_MODE=$MODE MDL_NODE_HOST=$HOST WALLET_DIR=$WALLET_DIR \
    go test ./src/api/integration/... -timeout=30s $VERBOSE $RUN_TESTS

API_FAIL=$?

fi

if [[ -z $TEST  || $TEST = "cli" ]]; then

# MDL_INTEGRATION_TESTS=1 MDL_INTEGRATION_TEST_MODE=$MODE RPC_ADDR=$RPC_ADDR \
#     go test ./src/cli/integration/... -timeout=30s $VERBOSE $RUN_TESTS

CLI_FAIL=$?

fi


echo "shutting down $COIN node"

# Shutdown mdl node
kill -s SIGINT $MDL_PID
wait $MDL_PID

rm "$BINARY"


if [[ (-z $TEST || $TEST = "api") && $API_FAIL -ne 0 ]]; then
  exit $API_FAIL
elif [[ (-z $TEST || $TEST = "cli") && $CLI_FAIL -ne 0 ]]; then
  exit $CLI_FAIL
else
  exit 0
fi
# exit $FAIL
