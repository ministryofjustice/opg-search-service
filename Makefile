all: go-lint gosec unit-test build scan swagger-generate down

.PHONY: build unit-test gosec swagger-generate swagger-up swagger-down docs

up:
	docker compose up -d --build search-service

go-lint:
	docker compose run --rm go-lint

test-results:
	mkdir -p -m 0777 test-results .gocache .trivy-cache

setup-directories: test-results

build:
	docker compose build search-service

unit-test: setup-directories
	docker compose up -d --wait postgres localstack
	docker compose run --rm test-runner
	docker compose down postgres localstack

scan: setup-directories
	docker compose run --rm trivy image --format table --exit-code 0 311462405659.dkr.ecr.eu-west-1.amazonaws.com/search_service:latest
	docker compose run --rm trivy image --format sarif --output /test-results/trivy.sarif --exit-code 1 311462405659.dkr.ecr.eu-west-1.amazonaws.com/search_service:latest

gosec: # Run Golang Security Checker
	docker compose run --rm gosec

swagger-generate: # Generate API swagger docs from inline code annotations using Go Swagger (https://goswagger.io/)
	docker compose run --rm swagger-generate

swagger-up docs: # Serve swagger API docs on port 8383
	docker compose up -d --force-recreate swagger-ui
	@echo "Swagger docs available on http://localhost:8383/"

down:
	docker compose down

provider-pact:
	PACT_HEADER="Authorization=Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzZXNzaW9uLWRhdGEiOiJzeXN0ZW0uYWRtaW5Ab3BndGVzdC5jb20iLCJpYXQiOjE2OTU4OTEwOTF9.vuuw3zE9sJYRUCNZyAdtksUsbSDJTrNKhQaL5Kvr34I" docker compose run --rm pact-verifier
