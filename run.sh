#!/usr/bin/env bash

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
echo "MDLlife binary dir:" "$DIR"
pushd "$DIR" >/dev/null
go run cmd/mdl/mdl.go --gui-dir="${DIR}/src/gui/static/" $@
popd >/dev/null
