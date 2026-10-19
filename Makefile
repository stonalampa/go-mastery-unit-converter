.PHONY: build run vet test lint vuln clean

GOVULNCHECK_VERSION := v1.9.0

build:
	go build -o bin/unit-converter ./cmd/unit-converter

vet:
	go vet ./...

test:
	go test -race ./...

lint:
	golangci-lint run

vuln:
	go run golang.org/x/vuln/cmd/govulncheck@$(GOVULNCHECK_VERSION) ./...

clean:
	rm -rf bin

run:
	go run ./cmd/unit-converter
