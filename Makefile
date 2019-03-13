GOFILES=$(shell find . -type f -name '*.go' -not -path "./vendor/*")
GOPACKAGES=$(shell go list ./... | grep -v /vendor/)

build: test
	GO111MODULE=on CGO_ENABLED=0 GOARCH=amd64 go build -a -installsuffix cgo -ldflags="-w -s" -o ./bin/updatey ./cmd/updatey

test:
	@go test ./... -coverprofile cover.out -race

fmt:
	@if [ -n "$$(gofmt -l ${GOFILES})" ]; then echo 'Please run gofmt -l -w on your code.' && exit 1; fi

lint:
	@golint ./pkg/...

vet:
	@go vet ./pkg/...