#!/usr/bin/env bash

PORT="46420"

echo "stopping any instances of mdl already running on this port $PORT"
lsof -i tcp:$PORT | awk 'NR!=1 {print $2}' | xargs kill -9  2>&1 >/dev/null | true