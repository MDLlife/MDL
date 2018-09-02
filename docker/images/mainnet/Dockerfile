# skycoin build
# reference https://github.com/skycoin/skycoin
ARG IMAGE_FROM=busybox
FROM golang:1.10.1-stretch AS build
ARG ARCH=amd64
ARG GOARM
ARG MDL_VERSION

COPY . $GOPATH/src/github.com/mdllife/mdl

# This code checks if MDL_VERSION is set and checkouts to that version if
# so. The git stash line prevents the build to fail if there are any uncommited
# changes in the working copy. It won't affect the host working copy.
RUN sh -c \
    'if test ${MDL_VERSION};then \
        echo "Revision is set to: "${MDL_VERSION}; \
        cd $GOPATH/src/github.com/mdllife/mdl; \
        git stash; \
        git checkout ${MDL_VERSION}; \
     fi'

RUN cd $GOPATH/src/github.com/mdllife/mdl && \
  CGO_ENABLED=0 GOOS=linux go install -a -installsuffix cgo ./...

# mdllife image
FROM alpine:3.7

ENV COIN="mdl" \
    RPC_ADDR="0.0.0.0:8330" \
    DATA_DIR="/data/.$COIN" \
    WALLET_DIR="/wallet" \
    USE_CSRF="1" \
    WALLET_NAME="$COIN_cli.wlt"

# copy all the binaries
COPY --from=build /go/bin/* /usr/bin/

# copy assets
COPY --from=build /go/src/github.com/mdllife/mdl/src/gui/static /usr/local/mdl/src/gui/static

# copy launcher
COPY docker_launcher.sh /usr/local/bin/docker_launcher.sh

#Copy certificates
COPY --from=build /etc/ssl/certs /etc/ssl/certs

# volumes
VOLUME $WALLET_DIR
VOLUME $DATA_DIR

EXPOSE 6000 8320 8330

ENTRYPOINT ["docker_launcher.sh", "--web-interface-addr=0.0.0.0", "--rpc-interface-addr=0.0.0.0", "--gui-dir=/usr/local/mdl/src/gui/static"]
