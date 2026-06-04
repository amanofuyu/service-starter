.PHONY: up down logs kafka-up kafka-down obs-up obs-down backup test vet fmt fmt-check compose-check check

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

vet:
	cd service && go vet ./...

fmt:
	cd service && gofmt -w .

fmt-check:
	cd service && test -z "$$(gofmt -l .)" || (echo "以下 Go 文件需要 gofmt:"; gofmt -l .; exit 1)

compose-check:
	docker compose --env-file .env.example config >/dev/null
	docker compose --env-file .env.example -f docker-compose.yml -f docker-compose.kafka.yml config >/dev/null
	docker compose --env-file .env.example -f docker-compose.yml -f docker-compose.obs.yml config >/dev/null

check: fmt-check test vet compose-check
