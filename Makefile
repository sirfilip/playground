.PHONY: build
build:
	go build .

.PHONY: test
test:
	go test -count=1 ./...

.PHONY: migrate
migrate:
	cd task_manager; goose up

.PHONY: migrate_reset
migrate_reset:
	cd task_manager; goose reset; goose up
