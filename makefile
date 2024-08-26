.PHONY: help
help:
	@echo Usage: 
	@echo run/api            -  run the api application
	@echo db/psql            -  connect to the db using psql
	@echo build/api          -  build the cmd/api application
	@echo audit              -  tidy dependencies and format, vet and test all code
	@echo db/migrations/up   -  run the up migrations using confirm as prerequisite
	@echo vendor             -  tidy and vendor dependencies

## build/api: run the cmd/api application
.PHONY: run/api
run/api:
	@echo 'Running cmd/api...'
	go run ./cmd/api

# db/psql: connect to the db using psql
.PHONY: db/psql
db/psql:
	psql ${AGGREGATE_DB_DSN}

# db/migrations/up: run the up migrations using confirm from confirm.p1 as prereq
.PHONY: db/migrations/up
db/migrations/up: confirm
	@echo Running up migrations...
	cd internal\sql\schema
	goose ${AGGREGATE_DB_DSN} up

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

## vendor: tidy and vendor dependencies
.PHONY: vendor
vendor:
	@echo 'Tidying and verifying module dependencies...'
	go mod tidy
	go mod verify
	@echo 'Vendoring dependencies...'
	go mod vendor

## build/api: build the cmd/api application
.PHONY: build/api
build/api:
	@echo 'Building cmd/api...'
	go build -ldflags '-s' -o ./bin/api.exe ./cmd/api
## For linux: GOOS=linux GOARCH=amd64 go build -ldflags='-s' -o bin/linux_amd64_api ./cmd/api