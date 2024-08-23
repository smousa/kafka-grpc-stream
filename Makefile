include .env

export USER_ID := $(shell id -u)
export GROUP_ID := $(shell id -g)

clean: export COMPOSE_PROFILES := ci
clean:
	docker compose down

lint:
	docker compose run --rm dev golangci-lint run

test:
	docker compose run --rm dev ginkgo -r --cover --label-filter="!e2e"

e2e-test:
	docker compose run --rm dev ginkgo -r --label-filter="e2e"

coverage:
	docker compose run --rm dev go-test-coverage --config=.testcoverage.yaml

mock:
	docker compose run --rm dev mockery

proto:
	docker compose run --rm dev protoc \
		--go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		protobuf/service.proto

version:
	@git diff --exit-code --quiet main VERSION || exit 0 && echo VERSION file not updated && exit 1
