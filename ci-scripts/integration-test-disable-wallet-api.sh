#!/bin/bash
# Runs "disable-wallet-api"-mode tests against a mdl node configured with the wallet API disabled.
# "disable-wallet-api"-mode confirms that no wallet related apis work, that the main index.html page
# does not load, and that a new wallet file is not created.

# Set Script Name variable
SCRIPT=`basename ${BASH_SOURCE[0]}`

# Find unused port
PORT="1024"
while $(lsof -Pi :$PORT -sTCP:LISTEN -t >/dev/null) ; do
    PORT=$((PORT+1))
done

COIN="${COIN:-mdl}"
RPC_PORT="$PORT"
HOST="http://127.0.0.1:$PORT"
RPC_ADDR="http://127.0.0.1:$RPC_PORT"
MODE="disable-wallet-api"
BINARY="${COIN}-integration-disable-wallet-api.test"
UPDATE=""
# run go test with -v flag
VERBOSE=""
# run go test with -run flag
RUN_TESTS=""
FAILFAST=""

usage () {
  echo "Usage: $SCRIPT"
  echo "Optional command line arguments"
  echo "-r <string>  -- Run test with -run flag"
  echo "-u <boolean> -- Update stable testdata"
  echo "-v <boolean> -- Run test with -v flag"
  echo "-f <boolean> -- Run test with -failfast flag"
  exit 1
}

while getopts "h?t:r:uvf" args; do
  case $args in
    h|\?)
        usage;
        exit;;
    r ) RUN_TESTS="-run ${OPTARG}";;
    u ) UPDATE="--update";;
    v ) VERBOSE="-v";;
    f ) FAILFAST="-failfast"
  esac
done

COVERAGEFILE="coverage/${BINARY}.coverage.out"
if [ -f "${COVERAGEFILE}" ]; then
    rm "${COVERAGEFILE}"
fi

set -euxo pipefail

CMDPKG=$(go list ./cmd/${COIN})
COVERPKG=$(dirname $(dirname ${CMDPKG}))
COMMIT=$(git rev-parse HEAD)
BRANCH=$(git rev-parse --abbrev-ref HEAD)
GOLDFLAGS="-X ${CMDPKG}.Commit=${COMMIT} -X ${CMDPKG}.Branch=${BRANCH}"

echo "checking if integration tests compile"
go test ./src/api/integration/...
go test ./src/cli/integration/...

DATA_DIR=$(mktemp -d -t ${COIN}-data-dir.XXXXXX)
WALLET_DIR="${DATA_DIR}/wallets"

if [[ ! "$DATA_DIR" ]]; then
  echo "Could not create temp dir"
  exit 1
fi

# Compile the mdl node
# We can't use "go run" because that creates two processes which doesn't allow us to kill it at the end
echo "compiling $COIN with coverage"
go test -c -ldflags "${GOLDFLAGS}" -tags testrunmain -o "$BINARY" -coverpkg="${COVERPKG}/..." ./cmd/${COIN}/

mkdir -p coverage/

# Run mdl node with pinned blockchain database
echo "starting $COIN node in background with http listener on $HOST"

./"$BINARY" -disable-networking=true \
            -genesis-signature eb10468d10054d15f2b6f8946cd46797779aa20a7617ceb4be884189f219bc9a164e56a5b9f7bec392a804ff3740210348d73db77a37adb542a8e08d429ac92700 \
            -genesis-address 2jBbGxZRGoQG1mqhPBnXnLTxK6oxsTf8os6 \
            -blockchain-public-key 0328c576d3f420e7682058a981173a4b374c7cc5ff55bf394d3cf57059bbe6456a \
            -db-path=./src/api/integration/testdata/blockchain-180.db \
            -web-interface-port=$PORT \
            -download-peerlist=false \
            -db-path=./src/api/integration/testdata/blockchain-180.db \
            -db-read-only=true \
            -launch-browser=false \
            -data-dir="$DATA_DIR" \
            -wallet-dir="$WALLET_DIR" \
            -enable-all-api-sets=true \
            -disable-api-sets=WALLET \
            -test.run "^TestRunMain$" \
            -test.coverprofile="${COVERAGEFILE}" \
            &
MDL_PID=$!

echo "$COIN node pid=$MDL_PID"

echo "sleeping for startup"
sleep 3
echo "done sleeping"

set +e

MDL_INTEGRATION_TESTS=1 MDL_INTEGRATION_TEST_MODE=$MODE MDL_NODE_HOST=$HOST WALLET_DIR=$WALLET_DIR \
    go test ./src/api/integration/... $FAILFAST $UPDATE -timeout=30s $VERBOSE $RUN_TESTS

FAIL=$?

echo "shutting down $COIN node"

# Shutdown mdl node
kill -s SIGINT $MDL_PID
wait $MDL_PID

rm "$BINARY"

exit $FAIL
