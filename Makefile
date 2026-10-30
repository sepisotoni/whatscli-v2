.PHONY: tidy build run test vet lint check install-test version

tidy:
	go mod tidy
build:
	CGO_ENABLED=0 go build -trimpath -o bin/whatshell ./cmd/whatshell
run:
	go run ./cmd/whatshell
test:
	go test -race -count=1 ./...
vet:
	go vet ./...
lint:
	golangci-lint run ./...
check:
	bash scripts/check.sh

install-test:
	bash scripts/test_install.sh
version:
	go run ./cmd/whatshell --version
