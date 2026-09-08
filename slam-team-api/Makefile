.PHONY: dev run build tidy vet test clean

dev:            ## Hot reload (requires: go install github.com/air-verse/air@latest)
	air

run:            ## Run without hot reload
	go run ./cmd/api

build:          ## Build production binary → bin/slam-team-api
	go build -o bin/slam-team-api ./cmd/api

tidy:
	go mod tidy

vet:
	go vet ./...

test:
	go test ./...

clean:
	rm -rf bin tmp build-errors.log
