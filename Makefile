infra-up:
	docker compose -f docker-compose.dev.yml up -d

infra-down:
	docker compose -f docker-compose.dev.yml down -v

server-up:
	go run cmd/server/main.go 