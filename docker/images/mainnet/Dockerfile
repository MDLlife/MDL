# skycoin build
# reference https://github.com/skycoin/skycoin
FROM golang:1.9 AS build

ENV MDL_VERSION=0.1.0

# dirs
RUN mkdir -p $GOPATH/src/github.com/mdllife/mdl

# clone
RUN wget -c https://github.com/mdllife/mdl/archive/v${MDL_VERSION}.tar.gz

# uncompress
RUN tar zxf v${MDL_VERSION}.tar.gz -C $GOPATH/src/github.com/mdllife/mdl/ --strip-components=1

# install cli
RUN cd $GOPATH/src/github.com/mdllife/mdl/cmd/cli && \
  CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /go/bin/skycoin-cli .

# install skycoin
RUN cd $GOPATH/src/github.com/mdllife/mdl/cmd/skycoin && \
  CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /go/bin/mdl .

# install address_gen
RUN cd $GOPATH/src/github.com/mdllife/mdl/cmd/address_gen && \
  CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /go/bin/address_gen .


# skycoin image
FROM alpine:3.7

ENV COIN="mdl" \
    RPC_ADDR="0.0.0.0:6430" \
    DATA_DIR="/root/.$COIN" \
    WALLET_DIR="/wallet" \
    WALLET_NAME="$COIN_cli.wlt"

# copy all the binaries
COPY --from=build /go/bin/* /usr/bin/

# copy assets
COPY --from=build /go/src/github.com/mdllife/mdl/src/gui/static /usr/local/mdl/static

# volumes
VOLUME $WALLET_DIR
VOLUME $DATA_DIR

EXPOSE 6000 6420 6430

CMD ["/usr/bin/mdl", "--gui-dir=/usr/local/mdl/static", "--web-interface-addr=0.0.0.0"]
