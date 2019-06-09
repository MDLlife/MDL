#!/usr/bin/env bash

# Runs mdl in desktop client configuration

set -x

PORT="8320"
IP_ADDR="127.0.0.1"

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
echo "mdl binary dir:" "$DIR"
pushd "$DIR" >/dev/null

COMMIT=$(git rev-parse HEAD)
BRANCH=$(git rev-parse --abbrev-ref HEAD)
GOLDFLAGS="${GOLDFLAGS} -X main.Commit=${COMMIT} -X main.Branch=${BRANCH}"

GORUNFLAGS=${GORUNFLAGS:-}

go run -ldflags "${GOLDFLAGS}" $GORUNFLAGS cmd/mdl/mdl.go \
    -web-interface=true \
    -web-interface-addr=${IP_ADDR} \
    -web-interface-port=${PORT} \
    -gui-dir="${DIR}/src/gui/static/" \
    -launch-browser=true \
    -enable-all-api-sets=true \
    -enable-gui=true \
    -log-level=debug \
    $@

popd >/dev/null
