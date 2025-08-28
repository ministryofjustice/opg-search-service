all: go-lint unit-test build scan swagger-generate down

.PHONY: build unit-test swagger-generate swagger-up swagger-down docs

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
	docker compose run --rm trivy image --format table --exit-code 0 311462405659.dkr.ecr.eu-west-1.amazonaws.com/sirius/search-service:latest
	docker compose run --rm trivy image --format sarif --output /test-results/trivy.sarif --exit-code 1 311462405659.dkr.ecr.eu-west-1.amazonaws.com/sirius/search-service:latest

swagger-generate: # Generate API swagger docs from inline code annotations using Go Swagger (https://goswagger.io/)
	docker compose run --rm swagger-generate

swagger-up docs: # Serve swagger API docs on port 8383
	docker compose up -d --force-recreate swagger-ui
	@echo "Swagger docs available on http://localhost:8383/"

down:
	docker compose down

# the docker command here generates an "Authorization=Bearer <jwt>" header so the pact verifier can talk to search-service
provider-pact-build:
	docker compose build pact-provider-state-api
provider-pact: pact_header = "$(shell docker compose run pact-provider-state-api python ./api/jwt_maker.py)"
provider-pact: provider-pact-build
	PACT_HEADER=${pact_header} docker compose run --rm pact-verifier
