#!/bin/bash
# Runs UI e2e tests against a mdl node configured with a pinned database

# Set Script Name variable
SCRIPT=`basename ${BASH_SOURCE[0]}`

# Find unused port
PORT="1024"
while $(lsof -Pi :$PORT -sTCP:LISTEN -t >/dev/null) ; do
    PORT=$((PORT+1))
done

RPC_ADDR="127.0.0.1:$PORT"
HOST="http://127.0.0.1:$PORT"
BINARY="mdl-integration"
E2E_PROXY_CONFIG=$(mktemp -t e2e-proxy.config.XXXXXX.js)

COMMIT=$(git rev-parse HEAD)
BRANCH=$(git rev-parse --abbrev-ref HEAD)
GOLDFLAGS="-X main.Commit=${COMMIT} -X main.Branch=${BRANCH}"

set -euxo pipefail

DATA_DIR=$(mktemp -d -t mdl-data-dir.XXXXXX)
WALLET_DIR="${DATA_DIR}/wallets"

if [[ ! "$DATA_DIR" ]]; then
  echo "Could not create temp dir"
  exit 1
fi

# Create a dummy wallet with an address existing in the blockchain-180.db dataset
mkdir "$WALLET_DIR"
cat >"${WALLET_DIR}/test_wallet.wlt" <<EOL
{
    "meta": {
        "coin": "mdl",
        "cryptoType": "scrypt-chacha20poly1305",
        "encrypted": "true",
        "filename": "test_wallet.wlt",
        "label": "Test wallet",
        "lastSeed": "",
        "secrets": "dgB7Im4iOjEwNDg1NzYsInIiOjgsInAiOjEsImtleUxlbiI6MzIsInNhbHQiOiIvelgxOFdPQUlzK1FQOXZZWi9aVXlDVktmZWMzY29UdjNzU2h6cENmWDNvPSIsIm5vbmNlIjoid0Qxb0U5VldycW9RTmJKVyJ9qFmBxQnP42SKJsQavIW/8chLo3alLx/KZI/lFFU96iZhTeSAfLNtPajX+4bcAdsdsPPhoBLNRBBuy1O2NImjZOVEc3YPCpXQO2Zj6/AZKu6zRldSSRbyk2blLngHr9Iv2oS4CcofCUdQF6tfc8soU/Vef9pZAHEUn0Soi1i9iprK3trkq0CfgP3LR3faltBfTkJCkOOjNGbHgDfZrGL6TZpllxjEAlO2jzYqMvmucowq3MDlTplFMJoE5Fvw47gjSuOpdRQ0yK4EgTabXKZJbbjvWZzE9pCYuUE=",
        "seed": "",
        "tm": "1529948542",
        "type": "deterministic",
        "version": "0.2"
    },
    "entries": [
        {
            "address": "R6aHqKWSQfvpdo2fGSrq4F1RYXkBWR9HHJ",
            "public_key": "03cef9d4635c6f075a479415805134daa1b5fda6e0f6a82b154e04b26db6afa770",
            "secret_key": ""
        }
    ]
}
EOL

# Compile the mdl node
# We can't use "go run" because that creates two processes which doesn't allow us to kill it at the end
echo "compiling mdl"
go build -o "$BINARY" -ldflags "${GOLDFLAGS}" cmd/mdl-cli/mdl.go

# Run mdl node with pinned blockchain database
echo "starting mdl node in background with http listener on $HOST"

./mdl-integration -disable-networking=true \
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
                      -enable-all-api-sets=true \
                      -enable-api-sets=INSECURE_WALLET_SEED \
                      -wallet-dir="$WALLET_DIR" \
                      &
MDL_PID=$!

echo "mdl node pid=$MDL_PID"

echo "sleeping for startup"
sleep 3
echo "done sleeping"

set +e


cat >$E2E_PROXY_CONFIG <<EOL
const PROXY_CONFIG = {
  "/api/*": {
    "target": "$HOST",
    "secure": false,
    "logLevel": "debug",
    "bypass": function (req) {
      req.headers["host"] = '$RPC_ADDR';
      req.headers["referer"] = '$HOST';
      req.headers["origin"] = '$HOST';
    }
  }
};
module.exports = PROXY_CONFIG;
EOL

# Run e2e tests
E2E_PROXY_CONFIG=$E2E_PROXY_CONFIG npm --prefix="./src/gui/static" run e2e-choose-config

RESULT=$?

echo "shutting down mdl node"

# Shutdown mdl node
kill -s SIGINT $MDL_PID
wait $MDL_PID

rm "$BINARY"
rm "$E2E_PROXY_CONFIG"

if [[ $RESULT -ne 0 ]]; then
  exit $RESULT
else
  exit 0
fi
