infra-up:
	docker compose -f docker-compose.dev.yml up -d

infra-down:
	docker compose -f docker-compose.dev.yml down -v

infra-reload: infra-down infra-up

server-up:
	go run cmd/server/main.go 