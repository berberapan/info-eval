include .envrc

# _______ HELPERS ___________

.PHONY: help	
## help: prints this message
help:
	@echo "Usage:"
	@sed -n "s/^##//p" ${MAKEFILE_LIST} | column -t -s ":" | sed -e "s/^/ /"

.PHONY: confirm
confirm:
	@echo -n "Are you sure [y/n] " && read ans && [ $${ans: -N} = y ]

# _______ DEV _______________

.PHONY: run/api
## run/api: runs the cmd/api application
run/api:
	@go run ./cmd/api -db-dsn=${DB_DSN}

.PHONY: db/psql
## db/psql: connect to the docker container with the database
db/psql:
	@docker exec -it info-eval psql ${DB_DSN}

# _______ MIGRATIONS ________

.PHONY: db/migrations/new
## db/migrations/new name=migration name: create new migration files. specifiy name at the end
db/migrations/new:
	@echo "Creating new migration files with name ${name}"
	migrate create -seq -ext=.sql -dir=./migrations/ ${name}

.PHONY: db/migrations/up
## db/migrations/up: apply all up database migrations
db/migrations/up: confirm
	@echo "Running up migrations.."
	@migrate -path ./migrations/ -database ${DB_DSN} up

# _______ QUALITY CONTROL ___

.PHONY: tidy
## tidy: tidy, verifying and vendoring module dependencies
tidy:
	@echo "Tidying module dependencies.."
	go mod tidy
	@echo "Verifying and vendoring.."
	go mod verify
	go mod vendor

.PHONY: test
## test: run tests
test:
	@echo "Running tests.."
	go test -v ./...

