.DEFAULT_GOAL := help
.PHONY: run run-help test test-386 test-amd64 check check-newcoin
.PHONY: integration-test-stable integration-test-stable-disable-csrf
.PHONY: integration-test-live integration-test-live-wallet
.PHONY: integration-test-disable-wallet-api integration-test-disable-seed-api
.PHONY: integration-test-enable-seed-api integration-test-enable-seed-api
.PHONY: integration-test-disable-gui integration-test-disable-gui
.PHONY: integration-test-db-no-unconfirmed integration-test-auth
.PHONY: install-linters format release clean-release clean-coverage
.PHONY: install-deps-ui build-ui build-ui-travis help newcoin merge-coverage
.PHONY: generate update-golden-files
.PHONY: fuzz-base58 fuzz-encoder

COIN ?= mdl

# Static files directory
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

GUI_STATIC_DIR = src/gui/static

PORT="8320"
RPC_PORT="8330"

IP_ADDR="127.0.0.1"

RPC_ADDR="$IP_ADDR:$RPC_PORT"

DATA_DIR=/tmp/tmp.Ict7RGVuhX
WALLET_DIR=$(DATA_DIR)/wallets

# Electron files directory
ELECTRON_DIR = electron

# ./src folder does not have code
# ./src/api folder does not have code
# ./src/util folder does not have code
# ./src/ciper/* are libraries manually vendored by cipher that do not need coverage
# ./src/gui/static* are static assets
# */testdata* folders do not have code
# ./src/consensus/example has no buildable code
PACKAGES = $(shell find ./src -type d -not -path '\./src' \
    							      -not -path '\./src/api' \
    							      -not -path '\./src/util' \
    							      -not -path '\./src/consensus/example' \
    							      -not -path '\./src/gui/static*' \
    							      -not -path '\./src/cipher/*' \
    							      -not -path '*/testdata*' \
    							      -not -path '*/test-fixtures*')

# Compilation output
BUILD_DIR = dist
BUILDLIB_DIR = $(BUILD_DIR)/libskycoin
LIB_DIR = lib
LIB_FILES = $(shell find ./lib/cgo -type f -name "*.go")
SRC_FILES = $(shell find ./src -type f -name "*.go")
BIN_DIR = bin
DOC_DIR = docs
INCLUDE_DIR = include
LIBSRC_DIR = lib/cgo
LIBDOC_DIR = $(DOC_DIR)/libc

# Compilation flags
CC_VERSION = $(shell $(CC) -dumpversion)
STDC_FLAG = $(python -c "if tuple(map(int, '$(CC_VERSION)'.split('.'))) < (6,): print('-std=C99'")
LIBC_LIBS = -lcriterion
LIBC_FLAGS = -I$(LIBSRC_DIR) -I$(INCLUDE_DIR) -I$(BUILD_DIR)/usr/include -L $(BUILDLIB_DIR) -L$(BUILD_DIR)/usr/lib

COMMIT=$(shell git rev-parse HEAD)
BRANCH=$(shell git rev-parse --abbrev-ref HEAD)
GOLDFLAGS="-X main.Commit=${COMMIT} -X main.Branch=${BRANCH}"

# Platform specific checks
OSNAME = $(TRAVIS_OS_NAME)

run-client:  ## Run mdl with desktop client configuration. To add arguments, do 'make ARGS="--foo" run'.
	./run-client.sh ${ARGS}

run-daemon:  ## Run mdl with server daemon configuration. To add arguments, do 'make ARGS="--foo" run'.
	./run-daemon.sh ${ARGS}

run:  ## Run the MDL node. To add arguments, do 'make ARGS="--foo" run'.
	go run cmd/mdl/mdl.go \
	    -web-interface=true \
        -web-interface-addr=${IP_ADDR} \
        -web-interface-port=${PORT} \
        -gui-dir="${DIR}/src/gui/static/" \
        -launch-browser=true \
        -enable-all-api-sets=true \
        -enable-gui=true \
        -verify-db=true \
        -log-level=debug \
        -download-peerlist=false \
        -disable-csrf=false \
        $@

run-help: ## Show MDL node help
	@go run cmd/mdl/mdl.go --help

run-integration-test-live: ## Run the mdl node configured for live integration tests
	./ci-scripts/run-live-integration-test-node.sh

run-integration-test-live-cover: ## Run the mdl node configured for live integration tests with coverage
	./ci-scripts/run-live-integration-test-node-cover.sh

test: ## Run tests for MDL
	@mkdir -p coverage/
	COIN=$(COIN) go test -coverpkg="github.com/MDLLife/MDL/..." -coverprofile=coverage/go-test-cmd.coverage.out -timeout=5m ./cmd/...
	COIN=$(COIN) go test -coverpkg="github.com/MDLLife/MDL/..." -coverprofile=coverage/go-test-src.coverage.out -timeout=5m ./src/...


test-386: ## Run tests for Skycoin with GOARCH=386
	GOARCH=386 COIN=$(COIN) go test ./cmd/... -timeout=5m
	GOARCH=386 COIN=$(COIN) go test ./src/... -timeout=5m

test-amd64: ## Run tests for MDL with GOARCH=amd64
	GOARCH=amd64 COIN=$(COIN) go test ./cmd/... -timeout=5m
	GOARCH=amd64 COIN=$(COIN) go test ./src/... -timeout=5m

configure-build:
	mkdir -p $(BUILD_DIR)/usr/tmp $(BUILD_DIR)/usr/lib $(BUILD_DIR)/usr/include
	mkdir -p $(BUILDLIB_DIR) $(BIN_DIR) $(INCLUDE_DIR)

$(BUILDLIB_DIR)/libskycoin.so: $(LIB_FILES) $(SRC_FILES)
	rm -Rf $(BUILDLIB_DIR)/libskycoin.so
	go build -buildmode=c-shared  -o $(BUILDLIB_DIR)/libskycoin.so $(LIB_FILES)
	mv $(BUILDLIB_DIR)/libskycoin.h $(INCLUDE_DIR)/

$(BUILDLIB_DIR)/libskycoin.a: $(LIB_FILES) $(SRC_FILES)
	rm -Rf $(BUILDLIB_DIR)/libskycoin.a
	go build -buildmode=c-archive -o $(BUILDLIB_DIR)/libskycoin.a  $(LIB_FILES)
	mv $(BUILDLIB_DIR)/libskycoin.h $(INCLUDE_DIR)/

build-libc-static: $(BUILDLIB_DIR)/libskycoin.a

build-libc-shared: $(BUILDLIB_DIR)/libskycoin.so

## Build libskycoin C client library
build-libc: configure-build build-libc-static build-libc-shared

## Build libskycoin C client library and executable C test suites
## with debug symbols. Use this target to debug the source code
## with the help of an IDE
build-libc-dbg: configure-build build-libc-static build-libc-shared
	$(CC) -g -o $(BIN_DIR)/test_libskycoin_shared $(LIB_DIR)/cgo/tests/*.c -lskycoin                    $(LDLIBS) $(LDFLAGS)
	$(CC) -g -o $(BIN_DIR)/test_libskycoin_static $(LIB_DIR)/cgo/tests/*.c $(BUILDLIB_DIR)/libskycoin.a $(LDLIBS) $(LDFLAGS)

test-libc: build-libc ## Run tests for libskycoin C client library
	echo "Compiling with $(CC) $(CC_VERSION) $(STDC_FLAG)"
	$(CC) -o $(BIN_DIR)/test_libskycoin_shared $(LIB_DIR)/cgo/tests/*.c $(LIB_DIR)/cgo/tests/testutils/*.c -lskycoin                    $(LDLIBS) $(LDFLAGS)
	$(CC) -o $(BIN_DIR)/test_libskycoin_static $(LIB_DIR)/cgo/tests/*.c $(LIB_DIR)/cgo/tests/testutils/*.c $(BUILDLIB_DIR)/libskycoin.a $(LDLIBS) $(LDFLAGS)
	$(LDPATHVAR)="$(LDPATH):$(BUILD_DIR)/usr/lib:$(BUILDLIB_DIR)" $(BIN_DIR)/test_libskycoin_shared
	$(LDPATHVAR)="$(LDPATH):$(BUILD_DIR)/usr/lib"                 $(BIN_DIR)/test_libskycoin_static

docs-libc:
	doxygen ./.Doxyfile
	moxygen -o $(LIBDOC_DIR)/API.md $(LIBDOC_DIR)/xml/

docs: docs-libc


lint: ## Run linters. Use make install-linters first.
	vendorcheck ./...
	golangci-lint run -c .golangci.yml ./...
	@# The govet version in golangci-lint is out of date and has spurious warnings, run it separately
	go vet -all ./...


check-newcoin: newcoin ## Check that make newcoin succeeds and no templated files are changed.
	@if [ "$(shell git diff ./cmd/mdl/mdl.go | wc -l | tr -d ' ')" != "0" ] ; then echo 'Changes detected after make newcoin' ; exit 2 ; fi
	@if [ "$(shell git diff ./cmd/mdl/mdl_test.go | wc -l | tr -d ' ')" != "0" ] ; then echo 'Changes detected after make newcoin' ; exit 2 ; fi
	@if [ "$(shell git diff ./src/params/params.go | wc -l | tr -d ' ')" != "0" ] ; then echo 'Changes detected after make newcoin' ; exit 2 ; fi

check: lint clean-coverage test test-386 \
	integration-test-stable integration-test-stable-disable-csrf \
	integration-test-disable-wallet-api integration-test-disable-seed-api \
	integration-test-enable-seed-api integration-test-disable-gui \
	integration-test-auth integration-test-db-no-unconfirmed ## Run tests and linters


integration-test-stable: ## Run stable integration tests
	COIN=$(COIN) ./ci-scripts/integration-test-stable.sh -c -x -n enable-csrf-header-check

integration-test-stable-disable-header-check: ## Run stable integration tests with header check disabled
	COIN=$(COIN) ./ci-scripts/integration-test-stable.sh -n disable-header-check

integration-test-stable-disable-csrf: ## Run stable integration tests with CSRF disabled
	COIN=$(COIN) ./ci-scripts/integration-test-stable.sh -n disable-csrf

integration-test-live: ## Run live integration tests
	COIN=$(COIN) ./ci-scripts/integration-test-live.sh -c

integration-test-live-wallet: ## Run live integration tests with wallet
	COIN=$(COIN) ./ci-scripts/integration-test-live.sh -w

integration-test-live-enable-header-check: ## Run live integration tests against a node with header check enabled
	COIN=$(COIN) ./ci-scripts/integration-test-live.sh

integration-test-live-disable-csrf: ## Run live integration tests against a node with CSRF disabled
	COIN=$(COIN) ./ci-scripts/integration-test-live.sh

integration-test-live-disable-networking: ## Run live integration tests against a node with networking disabled (requires wallet)
	COIN=$(COIN) ./ci-scripts/integration-test-live.sh -c -k

integration-test-disable-wallet-api: ## Run disable wallet api integration tests
	COIN=$(COIN) ./ci-scripts/integration-test-disable-wallet-api.sh

integration-test-enable-seed-api: ## Run enable seed api integration test
	COIN=$(COIN) ./ci-scripts/integration-test-enable-seed-api.sh

integration-test-disable-gui: ## Run tests with the GUI disabled
	COIN=$(COIN) ./ci-scripts/integration-test-disable-gui.sh

integration-test-db-no-unconfirmed: ## Run stable tests against the stable database that has no unconfirmed transactions
	COIN=$(COIN) ./ci-scripts/integration-test-stable.sh -d -n no-unconfirmed

integration-test-auth: ## Run stable tests with HTTP Basic auth enabled
	COIN=$(COIN) ./ci-scripts/integration-test-auth.sh

integration-test-server:
	rm -rf $(DATA_DIR);
	go build -ldflags ${GOLDFLAGS} -o /opt/gocode/src/github.com/MDLlife/MDL/mdl-integration \
	/opt/gocode/src/github.com/MDLlife/MDL/cmd/mdl/mdl.go;
	/opt/gocode/src/github.com/MDLlife/MDL/mdl-integration \
	-disable-networking=true \
	-genesis-signature eb10468d10054d15f2b6f8946cd46797779aa20a7617ceb4be884189f219bc9a164e56a5b9f7bec392a804ff3740210348d73db77a37adb542a8e08d429ac92700 \
	-genesis-address 2jBbGxZRGoQG1mqhPBnXnLTxK6oxsTf8os6 \
	-blockchain-public-key 0328c576d3f420e7682058a981173a4b374c7cc5ff55bf394d3cf57059bbe6456a \
	-peerlist-url https://downloads.mdl.net/blockchain/peers.txt \
	-web-interface-addr=$(IP_ADDR) \
	-web-interface-port=$(PORT) \
	-download-peerlist=false \
	-db-path=$(PWD)/src/api/integration/testdata/blockchain-180.db \
	-db-read-only=true \
	-enable-all-api-sets=true \
	-launch-browser=false \
	-data-dir=$(DATA_DIR) \
	-wallet-dir=$(WALLET_DIR) \
	-disable-csrf;

cover: ## Runs tests on ./src/ with HTML code coverage
	@echo "mode: count" > coverage-all.out
	$(foreach pkg,$(PACKAGES),\
		go test -coverprofile=coverage.out $(pkg);\
		tail -n +2 coverage.out >> coverage-all.out;)
	go tool cover -html=coverage-all.out

install-linters: ## Install linters
	go get -u github.com/FiloSottile/vendorcheck
	# For some reason this install method is not recommended, see https://github.com/golangci/golangci-lint#install
	# However, they suggest `curl ... | bash` which we should not do
	go get -u github.com/golangci/golangci-lint/cmd/golangci-lint

format: ## Formats the code. Must have goimports installed (use make install-linters).
	goimports -w -local github.com/MDLlife/mdl ./cmd
	goimports -w -local github.com/MDLlife/mdl ./src

install-deps-ui:  ## Install the UI dependencies
	cd $(GUI_STATIC_DIR) && npm install

lint-ui:  ## Lint the UI code
	cd $(GUI_STATIC_DIR) && npm run lint

test-ui:  ## Run UI tests
	cd $(GUI_STATIC_DIR) && npm run test

test-ui-e2e:  ## Run UI e2e tests
	./ci-scripts/ui-e2e.sh

clean-ui:  ## Builds the UI
	rm $(GUI_STATIC_DIR)/dist/* || true

build-ui:  ## Builds the UI
	cd $(GUI_STATIC_DIR) && npm run build

build-ui-travis:  ## Builds the UI for travis
	cd $(GUI_STATIC_DIR) && npm run build-travis

release: ## Build electron, standalone and daemon apps. Use osarch=${osarch} to specify the platform. Example: 'make release osarch=darwin/amd64', multiple platform can be supported in this way: 'make release osarch="darwin/amd64 windows/amd64"'. Supported architectures are: darwin/amd64 windows/amd64 windows/386 linux/amd64 linux/arm, the builds are located in electron/release folder.
	cd $(ELECTRON_DIR) && ./build.sh ${osarch}
	@echo release files are in the folder of electron/release

release-standalone: ## Build standalone apps. Use osarch=${osarch} to specify the platform. Example: 'make release-standalone osarch=darwin/amd64' Supported architectures are the same as 'release' command.
	cd $(ELECTRON_DIR) && ./build-standalone-release.sh ${osarch}
	@echo release files are in the folder of electron/release

release-electron: ## Build electron apps. Use osarch=${osarch} to specify the platform. Example: 'make release-electron osarch=darwin/amd64' Supported architectures are the same as 'release' command.
	cd $(ELECTRON_DIR) && ./build-electron-release.sh ${osarch}
	@echo release files are in the folder of electron/release

release-daemon: ## Build daemon apps. Use osarch=${osarch} to specify the platform. Example: 'make release-daemon osarch=darwin/amd64' Supported architectures are the same as 'release' command.
	cd $(ELECTRON_DIR) && ./build-daemon-release.sh ${osarch}
	@echo release files are in the folder of electron/release

release-cli: ## Build CLI apps. Use osarch=${osarch} to specify the platform. Example: 'make release-cli osarch=darwin/amd64' Supported architectures are the same as 'release' command.
	cd $(ELECTRON_DIR) && ./build-cli-release.sh ${osarch}
	@echo release files are in the folder of electron/release

clean-release: ## Remove all electron build artifacts
	rm -rf $(ELECTRON_DIR)/release
	rm -rf $(ELECTRON_DIR)/.gox_output
	rm -rf $(ELECTRON_DIR)/.daemon_output
	rm -rf $(ELECTRON_DIR)/.cli_output
	rm -rf $(ELECTRON_DIR)/.standalone_output
	rm -rf $(ELECTRON_DIR)/.electron_output

clean-coverage: ## Remove coverage output files
	rm -rf ./coverage/

newcoin: ## Rebuild cmd/$COIN/$COIN.go file from the template. Call like "make newcoin COIN=foo".
	go run cmd/newcoin/newcoin.go createcoin --coin $(COIN)

generate: ## Generate test interface mocks and struct encoders
	go generate ./src/...
	# mockery can't generate the UnspentPooler mock in package visor, patch it
	mv ./src/visor/blockdb/mock_unspent_pooler_test.go ./src/visor/mock_unspent_pooler_test.go
	sed -i "" -e 's/package blockdb/package visor/g' ./src/visor/mock_unspent_pooler_test.go
	sed -i "" -e 's/AddressHashes/blockdb.AddressHashes/g' ./src/visor/mock_unspent_pooler_test.go

install-generators: ## Install tools used by go generate
	go get github.com/vektra/mockery/.../
	go get github.com/skycoin/skyencoder/cmd/skyencoder

update-golden-files: ## Run integration tests in update mode
	./ci-scripts/integration-test-stable.sh -u >/dev/null 2>&1 || true
	./ci-scripts/integration-test-stable.sh -c -x -u >/dev/null 2>&1 || true
	./ci-scripts/integration-test-stable.sh -d -u >/dev/null 2>&1 || true
	./ci-scripts/integration-test-stable.sh -c -x -d -u >/dev/null 2>&1 || true

merge-coverage: ## Merge coverage files and create HTML coverage output. gocovmerge is required, install with `go get github.com/wadey/gocovmerge`
	@echo "To install gocovmerge do:"
	@echo "go get github.com/wadey/gocovmerge"
	gocovmerge coverage/*.coverage.out > coverage/all-coverage.merged.out
	go tool cover -html coverage/all-coverage.merged.out -o coverage/all-coverage.html
	@echo "Total coverage HTML file generated at coverage/all-coverage.html"
	@echo "Open coverage/all-coverage.html in your browser to view"

fuzz-base58: ## Fuzz the base58 package. Requires https://github.com/dvyukov/go-fuzz
	go-fuzz-build github.com/MDLlife/MDL/src/cipher/base58/internal
	go-fuzz -bin=base58fuzz-fuzz.zip -workdir=src/cipher/base58/internal

fuzz-encoder: ## Fuzz the encoder package. Requires https://github.com/dvyukov/go-fuzz
	go-fuzz-build github.com/MDLlife/MDL/src/cipher/encoder/internal
	go-fuzz -bin=encoderfuzz-fuzz.zip -workdir=src/cipher/encoder/internal

help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

release-all: ## clean and GUI_STATIC_DIR; clean and build release
	make clean-ui;
	make build-ui;
	make clean-release;
	make release;
