.PHONY: up down logs kafka-up kafka-down obs-up obs-down backup test fmt

up:
	docker compose up -d

down:
	docker compose down

logs:
	docker compose logs -f --tail=200

kafka-up:
	docker compose -f docker-compose.yml -f docker-compose.kafka.yml up -d

kafka-down:
	docker compose -f docker-compose.yml -f docker-compose.kafka.yml down

obs-up:
	docker compose -f docker-compose.yml -f docker-compose.obs.yml up -d

obs-down:
	docker compose -f docker-compose.yml -f docker-compose.obs.yml down

backup:
	docker compose --profile tools run --rm pg-backup

test:
	cd service && go test ./...

fmt:
	cd service && gofmt -w .
