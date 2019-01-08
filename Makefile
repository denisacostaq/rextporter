.DEFAULT_GOAL := help
.PHONY: test install-linters test-386 test-amd64 lint
 
build-grammar: ## Generate source code for REXT grammar
	nex -s src/rxt/grammar/lexer.nex

mocks: ## Create all mock files for unit tests
	echo "Generating mock files"
	mockery -all -dir ./src/config -output ./src/config/mocks

test-grammar: build-grammar ## Test cases for REXT lexer and parser
	go run cmd/rxtc/lexer.go < src/rxt/testdata/skyexample.rxt 2> src/rxt/testdata/skyexample.golden.orig
	diff -u src/rxt/testdata/skyexample.golden src/rxt/testdata/skyexample.golden.orig

test: mocks ## Run test with GOARCH=Default
	go test -count=1 github.com/simelo/rextporter/src/config
	go test -count=1 github.com/simelo/rextporter/src/scrapper
	go test -count=1 github.com/simelo/rextporter/src/memconfig

integration-test: ## Run integration tests with GOARCH=Default
	if ! screen -list | grep -q "fakeSkycoinForIntegrationTest"; then echo "creating screen fakeSkycoinForIntegrationTest"; screen -L -dm -S fakeSkycoinForIntegrationTest go run test/integration/fake_skycoin_node.go; else echo "fakeSkycoinForIntegrationTest screen already exist. quiting it to create a new one"; screen -S fakeSkycoinForIntegrationTest -X quit; screen -dm -S fakeSkycoinForIntegrationTest go run test/integration/fake_skycoin_node.go; fi
	sleep 3
	go test -count=1 -cpu=1 -parallel=1 github.com/simelo/rextporter/test/integration -args -test.v
	# screen -list can return a not 0 value, this is interpreted as a fail for travis, so use || true
	screen -list || true
	screen -S fakeSkycoinForIntegrationTest -X quit
	cat screenlog.0
	go test -cpu=1 -parallel=1  -count=1 github.com/simelo/rextporter/test/integration/skycoin

test-386: mocks ## Run tests  with GOARCH=386
	GOARCH=386 go test -count=1 github.com/simelo/rextporter/src/config
	GOARCH=386 go test -count=1 github.com/simelo/rextporter/src/scrapper
	GOARCH=386 go test -count=1 github.com/simelo/rextporter/src/memconfig

integration-test-386: ## Run integration tests with GOARCH=386
	if ! screen -list | grep -q "fakeSkycoinForIntegrationTest"; then echo "creating screen fakeSkycoinForIntegrationTest"; screen -L -dm -S fakeSkycoinForIntegrationTest go run test/integration/fake_skycoin_node.go; else echo "fakeSkycoinForIntegrationTest screen already exist. quiting it to create a new one"; screen -S fakeSkycoinForIntegrationTest -X quit; screen -dm -S fakeSkycoinForIntegrationTest go run test/integration/fake_skycoin_node.go; fi
	sleep 3
	GOARCH=386 go test -cpu=1 -parallel=1  -count=1 github.com/simelo/rextporter/test/integration -args -test.v
	# screen -list can return a not 0 value, this is interpreted as a fail for travis, so use || true
	screen -list || true
	screen -S fakeSkycoinForIntegrationTest -X quit
	cat screenlog.0
	GOARCH=386 go test -cpu=1 -parallel=1  -count=1 github.com/simelo/rextporter/test/integration/skycoin

test-amd64: mocks ## Run tests with GOARCH=amd64
	GOARCH=amd64 go test -count=1 github.com/simelo/rextporter/src/config
	GOARCH=amd64 go test -count=1 github.com/simelo/rextporter/src/scrapper
	GOARCH=amd64 go test -count=1 github.com/simelo/rextporter/src/memconfig

integration-test-amd64: ## Run integration tests with GOARCH=amd64
	if ! screen -list | grep -q "fakeSkycoinForIntegrationTest"; then echo "creating screen fakeSkycoinForIntegrationTest"; screen -L -dm -S fakeSkycoinForIntegrationTest go run test/integration/fake_skycoin_node.go; else echo "fakeSkycoinForIntegrationTest screen already exist. quiting it to create a new one"; screen -S fakeSkycoinForIntegrationTest -X quit; screen -dm -S fakeSkycoinForIntegrationTest go run test/integration/fake_skycoin_node.go; fi
	sleep 3
	GOARCH=amd64 go test -cpu=1 -parallel=1  -count=1 github.com/simelo/rextporter/test/integration -args -test.v
	# screen -list can return a not 0 value, this is interpreted as a fail for travis, so use || true
	screen -list || true
	screen -S fakeSkycoinForIntegrationTest -X quit
	cat screenlog.0
	GOARCH=amd64 go test -cpu=1 -parallel=1  -count=1 github.com/simelo/rextporter/test/integration/skycoin

lint: mocks ## Run linters. Use make install-linters first.
	ls src/
	vendorcheck ./...
	go vet -all ./...
	/tmp/bin/golangci-lint run -c .golangci.yml ./...

check:
	test

install-linters: ## Install linters
	go get -u github.com/FiloSottile/vendorcheck
	curl -sfL https://install.goreleaser.com/github.com/golangci/golangci-lint.sh | sh -s -- -b /tmp/bin v1.12.5

help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'