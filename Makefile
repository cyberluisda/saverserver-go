
define HELP_TEXT

	Makefile commands

	make distclean          - Delete all build artifacts

	make test               - Run the full test suite
	make test-coverage      - Run tests and make golang coverage reports

	make lint               - Run golang linter
	make lint-ci            - Run linter checks using golangci-lint tool (it must be installed before)
endef

help:
	$(info $(HELP_TEXT))

test: lint test-go

test-coverage: ./build/test
	go test ./server -coverprofile=./build/test/coverage.out
	go tool cover -html=./build/test/coverage.out -o ./build/test/coverage.html

lint:
	go vet ./server/...

lint-ci:
	golangci-lint run ./server/...

test-go:
	go test ./server/...

distclean:
	rm -fr ./build

./build/test:
	mkdir -p build/test
