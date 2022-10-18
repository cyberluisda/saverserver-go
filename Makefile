
define HELP_TEXT

	Makefile commands

	make distclean          - Delete all build artifacts

	make test               - Run the full test suite
	make test-coverage      - Run tests and make golang coverage reports

	make lint               - Run all linters


endef

help:
	$(info $(HELP_TEXT))

test: lint test-go

test-coverage: ./build/test
	go test ./server -coverprofile=./build/test/coverage.out
	go tool cover -html=./build/test/coverage.out -o ./build/test/coverage.html

lint:
	go vet ./server

test-go:
	go test ./server

distclelan:
	rm -fr ./build

./build/test:
	mkdir -p build/test
