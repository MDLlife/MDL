# MDL build binaries
# reference https://github.com/MDLLife/MDL
FROM golang:1.9-alpine AS build-go

COPY . $GOPATH/src/github.com/MDLLife/MDL

RUN cd $GOPATH/src/github.com/MDLLife/MDL && \
  CGO_ENABLED=0 GOOS=linux go install -a -installsuffix cgo ./...


# skycoin gui
FROM node:8.9 AS build-node

COPY . /mdl

# `unsafe` flag used as work around to prevent infinite loop in Docker
# see https://github.com/nodejs/node-gyp/issues/1236
RUN npm install -g --unsafe @angular/cli && \
    cd /mldl/src/gui/static && \
    yarn && \
    npm run build


# skycoin image
FROM alpine:3.7

ENV COIN="mdl" \
    RPC_ADDR="0.0.0.0:6430" \
    DATA_DIR="/data/.$COIN" \
    WALLET_DIR="/wallet" \
    WALLET_NAME="$COIN_cli.wlt"

RUN adduser -D mdl

USER skycoin

# copy binaries
COPY --from=build-go /go/bin/* /usr/bin/

# copy gui
COPY --from=build-node /mdl/src/gui/static /usr/local/mdl/src/gui/static

# volumes
VOLUME $WALLET_DIR
VOLUME $DATA_DIR

EXPOSE 6000 6420 6430

WORKDIR /usr/local/mdl

CMD ["mdl", "--web-interface-addr=0.0.0.0"]
