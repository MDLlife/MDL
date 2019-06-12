# mdl build
# reference https://github.com/MDLlife/MDL
ARG IMAGE_FROM=busybox
FROM golang:1.11-stretch AS build
ARG ARCH=amd64
ARG GOARM
ARG MDL_VERSION
ARG SCOMMIT
ARG SBRANCH
ARG STAG

COPY . $GOPATH/src/github.com/MDLlife/MDL

# This code checks if MDL_VERSION is set and checkouts to that version if
# so. The git stash line prevents the build to fail if there are any uncommited
# changes in the working copy. It won't affect the host working copy.
RUN sh -c \
    'if test ${MDL_VERSION};then \
        echo "Revision is set to: "${MDL_VERSION}; \
        cd $GOPATH/src/github.com/MDLlife/MDL; \
        git stash; \
        git checkout ${MDL_VERSION}; \
     fi'

ENV GOARCH="$ARCH" \
    GOARM="$GOARM" \
    CGO_ENABLED="0" \
    GOOS="linux" \
    GOLDFLAGS="-X main.Commit=${SCOMMIT} -X main.Branch=${SBRANCH}"

RUN cd $GOPATH/src/github.com/MDLlife/MDL && \
    echo "Building with GOLDFLAGS=$GOLDFLAGS GOARCH=$GOARCH GOARM=$GOARM CGO_ENABLED=$CGO_ENABLED GOOS=$GOOS " && \
    go install -ldflags "${GOLDFLAGS}" ./cmd/... && \
    sh -c "if test -d $GOPATH/bin/linux_arm ; then mv $GOPATH/bin/linux_arm/* $GOPATH/bin/; fi; \
           if test -d $GOPATH/bin/linux_arm64 ; then mv $GOPATH/bin/linux_arm64/* $GOPATH/bin/; fi"

RUN apt-get update && \
    apt-get install -y ca-certificates


RUN /bin/bash -c 'mkdir -p /tmp/files/{usr/bin,/usr/local/mdl/src/gui/static,/usr/local/bin/,/etc/ssl}'
RUN cp -r /go/bin/* /tmp/files/usr/bin/
RUN cp -r  /go/src/github.com/MDLlife/MDL/src/gui/static /tmp/files/usr/local/mdl/src/gui/static
RUN cp -r  /etc/ssl/certs /tmp/files/etc/ssl/certs
COPY docker_launcher.sh /tmp/files/usr/local/bin/docker_launcher.sh

# mdl image
FROM $IMAGE_FROM
ARG BDATE
ARG SCOMMIT
ARG SBRANCH
ARG STAG

# Image labels
LABEL "org.label-schema.name"="MDL" \
      "org.label-schema.description"="MDL core docker image" \
      "org.label-schema.vcs-url"="https://github.com/MDLlife/MDL/tree/develop/docker/images/mainnet" \
      "org.label-schema.vendor"="MDL project" \
      "org.label-schema.url"="mdl.net" \
      "org.label-schema.schema-version"="1.0" \
      "org.label-schema.build-date"=$BDATE \
      "org.label-schema.vcs-ref"=$SCOMMIT \
      "org.label-schema.version"=$STAG \
      "org.label-schema.usage"="https://github.com/MDLlife/MDL/blob/"$SCOMMIT"/docker/images/mainnet/README.md" \
      "org.label-schema.docker.cmd"="docker volume create mdl-data; docker volume create mdl-wallet; docker run -d -v mdl-data:/data/.mdl -v mdl-wallet:/wallet -p 6000:6000 -p 8320:8320 --name mdl-node-stable MDLlife/MDL"

ENV COIN="mdl"
ENV RPC_ADDR="http://0.0.0.0:8330" \
    DATA_DIR="/data/.$COIN" \
    WALLET_DIR="/wallet" \
    WALLET_NAME="$COIN_cli.wlt"

# copy all the binaries
COPY --from=build /tmp/files /

# volumes
VOLUME $WALLET_DIR
VOLUME $DATA_DIR

EXPOSE 6000 8320

ENTRYPOINT ["docker_launcher.sh", "--web-interface-addr=0.0.0.0", "--gui-dir=/usr/local/mdl/src/gui/static"]
