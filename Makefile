SOURCE_FILES?="./cmd/./... ./lib/./..."
BIN?=iriguchikun
TEST_PATTERN?=.
TEST_OPTIONS?=
OS=$(shell uname -s)

export PATH := ./bin:$(PATH)

# Install all the build and lint dependencies
setup:
	go get -u golang.org/x/tools/cmd/stringer
	go get -u golang.org/x/tools/cmd/cover
	curl -sfL https://install.goreleaser.com/github.com/golangci/golangci-lint.sh | sh
	curl -sfL https://install.goreleaser.com/github.com/caarlos0/bandep.sh | sh
ifeq ($(OS), Darwin)
	brew install dep
else
	curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh
endif
	dep ensure -vendor-only
	echo "make check" > .git/hooks/pre-commit
	chmod +x .git/hooks/pre-commit
.PHONY: setup

check:
	bandep --ban github.com/tj/assert
.PHONY: check

# Run all the tests
test:
	go test $(TEST_OPTIONS) -v -failfast -race -coverpkg=./... -covermode=atomic -coverprofile=coverage.txt $(SOURCE_FILES) -run $(TEST_PATTERN) -timeout=2m
.PHONY: cover

# Run all the tests and opens the coverage report
cover: test
	go tool cover -html=coverage.txt
.PHONY: cover

# gofmt and goimports all go files
fmt:
	find . -name '*.go' -not -wholename './vendor/*' | while read -r file; do gofmt -w -s "$$file"; goimports -w "$$file"; done
.PHONY: fmt

# Run all the linters
lint:
	./bin/golangci-lint run --tests=false --enable-all --disable=lll ./...
.PHONY: lint

# Run all the tests and code checks
ci: build test lint
.PHONY: ci

# Build a beta version of $(BIN)
build: $(BIN)
.PHONY: build

clean:
	rm $(BIN)
.PHONY: clean

$(BIN):
	go build -o $@ cmd/$(BIN)/main.go

## Generate the static documentation
#static:
#	@rm -rf dist/$(BIN).github.io
#	@mkdir -p dist
#	@git clone https://github.com/masahide/masahide.github.io.git dist/masahide.github.io
#	@rm -rf dist/masahide.github.io/theme
#	@static-docs \
#		--in docs \
#		--out dist/.github.io \
#		--title GoReleaser \
#		--subtitle "Deliver Go binaries as fast and easily as possible" \
#		--google UA-106198408-1
#.PHONY: static

# Show to-do items per file.
todo:
	@grep \
		--exclude-dir=vendor \
		--exclude-dir=node_modules \
		--exclude=Makefile \
		--text \
		--color \
		-nRo -E ' TODO:.*|SkipNow|nolint:.*' .
.PHONY: todo


.DEFAULT_GOAL := build
