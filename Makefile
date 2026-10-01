.PHONY: test-code run migrate-up migrate-down migrate-force

test-code:
	go run ./cmd/test-code

run:
	go run ./cmd/api

migrate-up:
	go run ./cmd/migrate -direction=up

migrate-down:
	go run ./cmd/migrate -direction=down

migrate-force:
	go run ./cmd/migrate -force=$(version)