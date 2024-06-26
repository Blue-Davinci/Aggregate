.PHONY: help
help:
	@echo Usage: 
	@echo build/api          -  build the cmd/api application
	@echo audit			     -  tidy dependencies and format, vet and test all code

## build/api: run the cmd/api application
.PHONY: run/api
run/api:
	@echo 'Running cmd/api...'
	go run ./cmd/api

## audit: tidy dependencies and format, vet and test all code
.PHONY: audit
audit:
	@echo 'Formatting code...'
	go fmt ./...
	@echo 'Vetting code...'
	go vet ./...
	staticcheck ./...
	@echo 'Running tests...'
	go test -race -vet=off ./...

## build/api: build the cmd/api application
.PHONY: build/api
build/api:
	@echo 'Building cmd/api...'
	go build -ldflags '-s' -o ./bin/api.exe ./cmd/api