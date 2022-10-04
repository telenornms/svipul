GIT_DESCRIBE:=$(shell git describe --always --tag --dirty)
VERSION_NO=$(shell echo ${GIT_DESCRIBE} | sed s/[v-]//g)
OS:=$(shell uname -s | tr A-Z a-z)
ARCH:=$(shell uname -m)

tpoll: $(wildcard *.go */*.go */*/*.go)
	@echo 🤸 go build !
	@go build -ldflags "-X main.versionNo=${VERSION_NO}" -o tpoll ./cmd/tpoll

all: tpoll

check: test fmtcheck vet

# Can't for the life of me remember where this came from and it's seemingly
# gone now, so removed from check.
lint:
	@echo 🐉 Linting code
	@golint -set_exit_status

vet:
	@echo 🔬 Vetting code
	@go vet ./...

fmtcheck:
	@echo 🦉 Checking format with gofmt -d -s
	@if [ "x$$(find . -name '*.go' -not -wholename './gen/*' -and -not -wholename './vendor/*' -exec gofmt -d -s {} +)" != "x" ]; then find . -name '*.go' -not -wholename './gen/*' -and -not -wholename './vendor/*' -exec gofmt -d -s {} +; exit 1; fi

fmtfix:
	@echo 🎨 Fixing formating
	@find . -name '*.go' -not -wholename './gen/*' -and -not -wholename './vendor/*' -exec gofmt -d -s -w {} +

test:
	@echo 🧐 Testing, without SQL-tests
	@go test -short ./...

bench:
	@echo 🏋 Benchmarking
	@go test -run ^Bench -benchtime 1s -bench Bench ./... | grep Benchmark

covergui:
	@echo 🧠 Testing, with coverage analysis
	@go test -short -coverpkg ./... -covermode=atomic -coverprofile=coverage.out ./...
	@echo 💡 Generating HTML coverage report and opening browser
	@go tool cover -html coverage.out

.PHONY: clean test bench help install rpm release
