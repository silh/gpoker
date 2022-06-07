.PHONY: build
build:
	go build -o gpoker ./cmd/gpoker

.PHONY: test
test:
	go test ./... -race
