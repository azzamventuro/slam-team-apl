.PHONY: dev run build build-cli tidy vet test clean migrate-up migrate-down migrate-version seed superadmin

dev:            ## Hot reload (requires: go install github.com/air-verse/air@latest)
	air

run:            ## Run without hot reload
	go run ./cmd/api

build:          ## Build production binary → bin/slam-team-api
	go build -o bin/slam-team-api ./cmd/api

build-cli:      ## Build operational CLI → bin/slamctl
	go build -o bin/slamctl ./cmd/slamctl

tidy:
	go mod tidy

vet:
	go vet ./...

test:
	go test ./...

clean:
	rm -rf bin tmp build-errors.log

migrate-up:     ## Apply pending SQL migrations
	go run ./cmd/slamctl migrate up

migrate-down:   ## Roll back one migration (make migrate-down N=3, or N=all)
	go run ./cmd/slamctl migrate down $(N)

migrate-version: ## Show the current schema version
	go run ./cmd/slamctl migrate version

seed:           ## Run idempotent reference-data seeders (make seed NAME=rbac)
	go run ./cmd/slamctl seed $(NAME)

superadmin:     ## Bootstrap the first super admin (see: go run ./cmd/slamctl create-superadmin -h)
	go run ./cmd/slamctl create-superadmin $(ARGS)
