.PHONY: build
build:
	go build -o gpoker ./...

.PHONY: test
test:
	go test ./... -race
