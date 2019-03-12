.DEFAULT_GOAL := help
.PHONY: run run-help test test-core test-libc test-lint build-libc check cover
.PHONY: integration-test-stable integration-test-stable-disable-csrf
.PHONY: integration-test-live integration-test-live-wallet
.PHONY: integration-test-disable-wallet-api integration-test-disable-seed-api
.PHONY: integration-test-enable-seed-api integration-test-enable-seed-api
.PHONY: integration-test-disable-gui integration-test-disable-gui
.PHONY: install-linters format release clean-release install-deps-ui build-ui help

# Static files directory
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

GUI_STATIC_DIR = src/gui/static

PORT="8320"
RPC_PORT="8330"

IP_ADDR="127.0.0.1"

RPC_ADDR="$IP_ADDR:$RPC_PORT"

DATA_DIR=$(mktemp -d -t mdl-data-dir.XXXXXX)
WALLET_DIR="${DATA_DIR}/wallets"

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

# Platform specific checks
OSNAME = $(TRAVIS_OS_NAME)

ifeq ($(shell uname -s),Linux)
  LDLIBS=$(LIBC_LIBS) -lpthread
  LDPATH=$(shell printenv LD_LIBRARY_PATH)
  LDPATHVAR=LD_LIBRARY_PATH
  LDFLAGS=$(LIBC_FLAGS) $(STDC_FLAG)
ifndef OSNAME
  OSNAME = linux
endif
else ifeq ($(shell uname -s),Darwin)
ifndef OSNAME
  OSNAME = osx
endif
  LDLIBS = $(LIBC_LIBS)
  LDPATH=$(shell printenv DYLD_LIBRARY_PATH)
  LDPATHVAR=DYLD_LIBRARY_PATH
  LDFLAGS=$(LIBC_FLAGS) -framework CoreFoundation -framework Security
else
  LDLIBS = $(LIBC_LIBS)
  LDPATH=$(shell printenv LD_LIBRARY_PATH)
  LDPATHVAR=LD_LIBRARY_PATHTestUnspentMaybeBuildIndexesNoIndexNoHead
  LDFLAGS=$(LIBC_FLAGS)
endif

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
        -rpc-interface=true \
        -log-level=debug \
        -download-peerlist=false \
        -disable-csrf=false \
        $@

run-help: ## Show MDL node help
	@go run cmd/mdl/mdl.go --help

test: ## Run tests for MDL
	go test ./cmd/... -timeout=5m
	go test ./src/api/... -timeout=5m
	go test ./src/cipher/... -timeout=5m
	go test ./src/cli/... -timeout=5m
	go test ./src/coin/... -timeout=5m
	go test ./src/consensus/... -timeout=5m
	go test ./src/daemon/... -timeout=5m
	go test ./src/mdl/... -timeout=5m
	go test ./src/util/... -timeout=5m
	go test ./src/visor/... -timeout=5m
	go test ./src/wallet/... -timeout=5m

test-386: ## Run tests for MDL with GOARCH=386
	GOARCH=386 go test ./cmd/... -timeout=5m
	GOARCH=386 go test ./src/... -timeout=5m

test-amd64: ## Run tests for MDL with GOARCH=amd64
	GOARCH=amd64 go test ./cmd/... -timeout=5m
	GOARCH=amd64 go test ./src/... -timeout=5m

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
	golangci-lint run --no-config  --deadline=3m --concurrency=2 --disable-all --tests --skip-dirs=lib/cgo \
		-E goimports \
		-E golint \
		-E varcheck \
		-E unparam \
		-E structcheck \
		./...
	# lib cgo can't use golint because it needs export directives in function docstrings that do not obey golint rules
	golangci-lint run --no-config  --deadline=3m --concurrency=2 --disable-all --tests \
		-E goimports \
		-E varcheck \
		-E unparam \
		-E structcheck \
		./lib/cgo/...

check: lint test integration-test-stable integration-test-stable-disable-csrf integration-test-disable-wallet-api integration-test-disable-seed-api integration-test-enable-seed-api integration-test-disable-gui ## Run tests and linters


integration-test-stable: ## Run stable integration tests
	GOCACHE=off COIN=$(COIN) ./ci-scripts/integration-test-stable.sh -c -n enable-csrf

integration-test-stable-disable-csrf: ## Run stable integration tests with CSRF disabled
	GOCACHE=off COIN=$(COIN) ./ci-scripts/integration-test-stable.sh -n disable-csrf

integration-test-live: ## Run live integration tests
	GOCACHE=off COIN=$(COIN) ./ci-scripts/integration-test-live.sh -c

integration-test-live-wallet: ## Run live integration tests with wallet
	GOCACHE=off COIN=$(COIN) ./ci-scripts/integration-test-live.sh -w

integration-test-live-disable-csrf: ## Run live integration tests against a node with CSRF disabled
	GOCACHE=off COIN=$(COIN) ./ci-scripts/integration-test-live.sh

integration-test-live-disable-networking: ## Run live integration tests against a node with networking disabled (requires wallet)
	GOCACHE=off COIN=$(COIN) ./ci-scripts/integration-test-live.sh -c -k

integration-test-disable-wallet-api: ## Run disable wallet api integration tests
	GOCACHE=off COIN=$(COIN) ./ci-scripts/integration-test-disable-wallet-api.sh

integration-test-enable-seed-api: ## Run enable seed api integration test
	GOCACHE=off COIN=$(COIN) ./ci-scripts/integration-test-enable-seed-api.sh

integration-test-disable-gui: ## Run tests with the GUI disabled
	GOCACHE=off COIN=$(COIN) ./ci-scripts/integration-test-disable-gui.sh

integration-test-db-no-unconfirmed: ## Run stable tests against the stable database that has no unconfirmed transactions
	GOCACHE=off COIN=$(COIN) ./ci-scripts/integration-test-stable.sh -d -n no-unconfirmed

integration-test-auth: ## Run stable tests with HTTP Basic auth enabled
	GOCACHE=off COIN=$(COIN) ./ci-scripts/integration-test-auth.sh

integration-test-server:
	go build -o /opt/gocode/src/github.com/MDLlife/MDL/mdl-integration \
	/opt/gocode/src/github.com/MDLlife/MDL/cmd/mdl/mdl.go;
	/opt/gocode/src/github.com/MDLlife/MDL/mdl-integration \
	-disable-networking=true \
	-genesis-signature eb10468d10054d15f2b6f8946cd46797779aa20a7617ceb4be884189f219bc9a164e56a5b9f7bec392a804ff3740210348d73db77a37adb542a8e08d429ac92700 \
	-genesis-address 2jBbGxZRGoQG1mqhPBnXnLTxK6oxsTf8os6 \
	-blockchain-public-key 0328c576d3f420e7682058a981173a4b374c7cc5ff55bf394d3cf57059bbe6456a \
	-db-path=./src/api/integration/testdata/blockchain-180.db \
	-peerlist-url https://downloads.mdl.net/blockchain/peers.txt \
	-web-interface-addr=$(IP_ADDR) \
	-web-interface-port=$(PORT) \
	-download-peerlist=false \
	-db-path=./src/api/integration/testdata/blockchain-180.db \
	-db-read-only=true \
	-rpc-interface=true \
	-enable-all-api-sets=true \
	-launch-browser=false \
	-data-dir="$DATA_DIR" \
	-wallet-dir="$WALLET_DIR" \
	-disable-csrf;

cover: ## Runs tests on ./src/ with HTML code coverage
	@echo "mode: count" > coverage-all.out
	$(foreach pkg,$(PACKAGES),\
		go test -coverprofile=coverage.out $(pkg);\
		tail -n +2 coverage.out >> coverage-all.out;)
	go tool cover -html=coverage-all.out

install-linters: ## Install linters
	go get -u github.com/FiloSottile/vendorcheck
	go get -u github.com/golangci/golangci-lint/cmd/golangci-lint

install-deps-libc: configure-build ## Install locally dependencies for testing libskycoin
	git clone --recursive https://github.com/skycoin/Criterion $(BUILD_DIR)/usr/tmp/Criterion
	mkdir $(BUILD_DIR)/usr/tmp/Criterion/build
	cd    $(BUILD_DIR)/usr/tmp/Criterion/build && cmake .. && cmake --build .
	mv    $(BUILD_DIR)/usr/tmp/Criterion/build/libcriterion.* $(BUILD_DIR)/usr/lib/
	cp -R $(BUILD_DIR)/usr/tmp/Criterion/include/* $(BUILD_DIR)/usr/include/

format:  # Formats the code. Must have goimports installed (use make install-linters).
	goimports -w -local github.com/mdllife/mdl ./cmd
	goimports -w -local github.com/mdllife/mdl ./src
	goimports -w -local github.com/mdllife/mdl ./lib

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

release: ## Build electron apps, the builds are located in electron/release folder.
	cd $(ELECTRON_DIR) && yarn && ./build.sh
	@echo release files are in the folder of electron/release

clean-release: ## Clean dist files and delete all builds in electron/release
	rm $(ELECTRON_DIR)/release/* || true

help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

release-all: ## clean and GUI_STATIC_DIR; clean and build release
	make clean-ui;
	make build-ui;
	make clean-release;
	make release;