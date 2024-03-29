.PHONY: build
build:
	go build -o gpoker ./cmd/gpoker

.PHONY: test
test:
	@go test ./... -race -coverpkg=./cmd/...,./pkg/... -cover -coverprofile coverage.out
	@go tool cover -func coverage.out

.PHONY: run
run:
	go run ./cmd/gpoker
