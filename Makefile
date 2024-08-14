include .env

clean: export COMPOSE_PROFILES := ci
clean:
	docker compose down

lint:
	docker compose run --rm dev golangci-lint run

test:
	docker compose run --rm dev ginkgo -r --cover

coverage:
	docker compose run --rm dev go-test-coverage --config=.testcoverage.yaml

version:
	test ! `git diff main --exit-code --quiet VERSION`
