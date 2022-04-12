.PHONY: test test-up test-run test-down build go-test gosec swagger-generate swagger-up swagger-down docs
test: test-up test-run test-down

test-up:
	docker-compose --project-name search-service-test up -d localstack
	docker-compose --project-name search-service-test run --rm wait-for-it -address=localstack:4571 --timeout=30
	docker-compose --project-name search-service-test -f docker-compose.yml -f docker-compose.test.yml up -d postgres
	docker-compose --project-name search-service-test run --rm wait-for-it -address=postgres:5432 --timeout=30

test-run:
	docker-compose --project-name search-service-test -f docker-compose.yml -f docker-compose.test.yml run --rm search_service_test make go-test

test-down:
	docker-compose --project-name search-service-test down

build:
	docker-compose build search_service

go-test:
	go mod download
	gotestsum --format short-verbose -- -coverprofile=../cover.out ./...

gosec: # Run Golang Security Checker
	docker-compose --project-name search-service-gosec -f docker-compose.yml -f docker-compose.test.yml run --rm search_service_gosec

swagger-generate: # Generate API swagger docs from inline code annotations using Go Swagger (https://goswagger.io/)
	docker-compose --project-name search-service-docs-generate \
    -f docker-compose.yml run --rm swagger-generate
	docker-compose --project-name search-service-docs-generate down

swagger-up: # Serve swagger API docs on port 8383
	docker-compose --project-name search-service-docs \
    -f docker-compose.yml up -d --force-recreate swagger-ui

swagger-down:
	docker-compose --project-name search-service-docs down

docs: # Alias for make swagger-up (Serve API swagger docs)
	make swagger-up
