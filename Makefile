.PHONY: test test-go test-python up down logs

test: test-go test-python

test-go:
	cd backend && go test ./...

test-python:
	cd worker && python -m pytest -q

up:
	docker compose up --build

down:
	docker compose down -v

logs:
	docker compose logs -f api worker
