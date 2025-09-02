.PHONY: build build-alpine clean test test-unit test-integration test-coverage test-postgres test-all help default docker-build

BIN_NAME=bashhub-server
VERSION=$(shell git tag | sort --version-sort -r | head -1)
GIT_COMMIT=$(shell git rev-parse HEAD)
GIT_DIRTY=$(shell test -n "`git status --porcelain`" && echo "+CHANGES" || true)
BUILD_DATE=$(shell date '+%Y-%m-%d-%H:%M:%S')
IMAGE_NAME="pedromol/bashhub-server"


default: help

help:
	@echo 'Management commands for bashhub-server:'
	@echo
	@echo 'Usage:'
	@echo '    make build           Compile the project'
	@echo '    make docker-build    Build docker image'
	@echo '    make clean           Clean the directory tree'
	@echo '    make test            Run all tests'
	@echo '    make test-unit       Run unit tests only'
	@echo '    make test-integration Run integration tests only'
	@echo '    make test-coverage   Run tests with coverage report'
	@echo '    make test-sql-injection Run SQL injection prevention tests'
	@echo '    make test-security   Run all security-related tests'
	@echo '    make test-postgres   Start postgres in ephemeral docker container and run backend tests'
	@echo '    make test-all        Run all test suites'
	@echo
	@echo 'Advanced testing with scripts/run_tests.sh:'
	@echo '    ./scripts/run_tests.sh -t unit           Run unit tests only'
	@echo '    ./scripts/run_tests.sh -t integration    Run integration tests only'
	@echo '    ./scripts/run_tests.sh -c               Run with coverage report'
	@echo '    ./scripts/run_tests.sh -p               Run PostgreSQL integration tests'
	@echo

build:
	@echo "building $(BIN_NAME) $(VERSION)"
	@echo "GOPATH=$(GOPATH)"
	go build  -ldflags "-X github.com/pedromol/bashhub-server/cmd.Version=$(VERSION) -X github.com/pedromol/bashhub-server/cmd.GitCommit=$(GIT_COMMIT) -X github.com/pedromol/bashhub-server/cmd.BuildDate=$(BUILD_DATE)" -o bin/${BIN_NAME}

docker-build:
	docker build --no-cache=true --build-arg VERSION=${VERSION} --build-arg BUILD_DATE=${BUILD_DATE} --build-arg GIT_COMMIT=${GIT_COMMIT}  -t $(IMAGE_NAME) .

clean:
	@test ! -e bin/$(BIN_NAME) || rm bin/$(BIN_NAME)

test: test-unit test-integration

test-unit:
	@echo "Running unit tests..."
	go test -v ./internal/...

test-integration:
	@echo "Running integration tests..."
	go test -v -tags=integration ./...

test-sql-injection:
	@echo "Running SQL injection prevention tests..."
	go test -v ./internal/db/... -run "TestSQLInjection"

test-security:
	@echo "Running security-related tests..."
	make test-sql-injection

test-coverage:
	@echo "Running tests with coverage..."
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

test-postgres:
	@echo "Starting PostgreSQL for integration tests..."
	./scripts/run_tests.sh -p -t integration

test-all: test test-postgres



