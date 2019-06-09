#!/bin/bash
# Runs "stable"-mode tests against a mdl node configured with a pinned database
# "stable" mode tests assume the blockchain data is static, in order to check API responses more precisely
# $TEST defines which test to run i.e, cli or api; If empty both are run

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
MODE="stable"
NAME=""
TEST=""
UPDATE=""
# run go test with -v flag
VERBOSE=""
# run go test with -run flag
RUN_TESTS=""
DISABLE_CSRF="-disable-csrf"
USE_CSRF=""
DISABLE_HEADER_CHECK=""
HEADER_CHECK="1"
DB_NO_UNCONFIRMED=""
DB_FILE="blockchain-180.db"

usage () {
  echo "Usage: $SCRIPT"
  echo "Optional command line arguments"
  echo "-t <string>  -- Test to run, api or cli; empty runs both tests"
  echo "-r <string>  -- Run test with -run flag"
  echo "-n <string>  -- Specific name for this test, affects coverage output files"
  echo "-u <boolean> -- Update stable testdata"
  echo "-v <boolean> -- Run test with -v flag"
  echo "-c <boolean> -- Run tests with CSRF enabled"
  echo "-x <boolean> -- Run test with header check disabled"
  echo "-d <boolean> -- Run tests without unconfirmed transactions"
  exit 1
}

while getopts "h?t:r:n:uvcxd" args; do
  case $args in
    h|\?)
        usage;
        exit;;
    t ) TEST=${OPTARG};;
    r ) RUN_TESTS="-run ${OPTARG}";;
    n ) NAME="-${OPTARG}";;
    u ) UPDATE="--update";;
    v ) VERBOSE="-v";;
    d ) DB_NO_UNCONFIRMED="1"; DB_FILE="blockchain-180-no-unconfirmed.db";;
    c ) DISABLE_CSRF=""; USE_CSRF="1";;
    x ) DISABLE_HEADER_CHECK="-disable-header-check"; HEADER_CHECK="";
  esac
done

BINARY="${COIN}-integration${NAME}.test"

COVERAGEFILE="coverage/${BINARY}.coverage.out"
if [ -f "${COVERAGEFILE}" ]; then
    rm "${COVERAGEFILE}"
fi

set -euxo pipefail

COMMIT=$(git rev-parse HEAD)
BRANCH=$(git rev-parse --abbrev-ref HEAD)
CMDPKG=$(go list ./cmd/${COIN})
COVERPKG=$(dirname $(dirname ${CMDPKG}))
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
echo "starting $COIN node in background with http listener on $HOST, WALLET_DIR=${WALLET_DIR}"

./"$BINARY" -disable-networking=true \
            -genesis-signature eb10468d10054d15f2b6f8946cd46797779aa20a7617ceb4be884189f219bc9a164e56a5b9f7bec392a804ff3740210348d73db77a37adb542a8e08d429ac92700 \
            -genesis-address 2jBbGxZRGoQG1mqhPBnXnLTxK6oxsTf8os6 \
            -blockchain-public-key 0328c576d3f420e7682058a981173a4b374c7cc5ff55bf394d3cf57059bbe6456a \
            -web-interface-port=${PORT} \
            -download-peerlist=false \
            -db-path=$PWD/src/api/integration/testdata/${DB_FILE} \
            -db-read-only=true \
            -launch-browser=false \
            -data-dir="${DATA_DIR}" \
            -enable-all-api-sets=true \
            -wallet-dir="${WALLET_DIR}" \
            $DISABLE_CSRF \
            $DISABLE_HEADER_CHECK \
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

MDL_INTEGRATION_TESTS=1 MDL_INTEGRATION_TEST_MODE=$MODE MDL_NODE_HOST=$HOST \
	USE_CSRF=$USE_CSRF HEADER_CHECK=$HEADER_CHECK DB_NO_UNCONFIRMED=$DB_NO_UNCONFIRMED COIN=$COIN \
    go test ./src/api/integration/... $UPDATE -timeout=3m $VERBOSE $RUN_TESTS

API_FAIL=$?

fi

if [[ -z $TEST  || $TEST = "cli" ]]; then

MDL_INTEGRATION_TESTS=1 MDL_INTEGRATION_TEST_MODE=$MODE RPC_ADDR=$RPC_ADDR \
	USE_CSRF=$USE_CSRF HEADER_CHECK=$HEADER_CHECK DB_NO_UNCONFIRMED=$DB_NO_UNCONFIRMED COIN=$COIN \
    go test ./src/cli/integration/... $UPDATE -timeout=3m $VERBOSE $RUN_TESTS

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
