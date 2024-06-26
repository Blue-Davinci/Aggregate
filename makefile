.PHONY: help
help:
	@echo Usage: 
	@echo build/api          -  build the cmd/api application

## build/api: run the cmd/api application
.PHONY: run/api
run/api:
	@echo 'Running cmd/api...'
	go run ./cmd/api

## build/api: build the cmd/api application
.PHONY: build/api
build/api:
	@echo 'Building cmd/api...'
	go build -ldflags '-s' -o ./bin/api.exe ./cmd/api