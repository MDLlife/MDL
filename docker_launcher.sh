#!/bin/sh
COMMAND="mdl --data-dir $DATA_DIR --wallet-dir $WALLET_DIR $@"

adduser -D -u 10000 mdl

if [[ \! -d $DATA_DIR ]]; then
    mkdir -p $DATA_DIR
fi
if [[ \! -d $WALLET_DIR ]]; then
    mkdir -p $WALLET_DIR
fi

chown -R mdl:mdl $( realpath $DATA_DIR )
chown -R mdl:mdl $( realpath $WALLET_DIR )

su mdl -c "$COMMAND"
