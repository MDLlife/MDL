#!/usr/bin/env bash

# Runs the node with configuration necessary for running the live integration tests

set -exu

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
pushd "$DIR/.." >/dev/null

COIN=${COIN:-mdl}
COMMIT=$(git rev-parse HEAD)
BRANCH=$(git rev-parse --abbrev-ref HEAD)
GOLDFLAGS="-X main.Commit=${COMMIT} -X main.Branch=${BRANCH}"

go run -ldflags "${GOLDFLAGS}" cmd/mdl/mdl.go \
    -gui-dir="${DIR}/src/gui/static/" \
    -launch-browser=false \
    -enable-all-api-sets=true \
    -enable-api-sets=DEPRECATED_WALLET_SPEND,INSECURE_WALLET_SEED \
    $@

popd >/dev/null
