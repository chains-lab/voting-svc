DB_URL=postgresql://postgres:postgres@localhost:7000/postgres?sslmode=disable
OPENAPI_GENERATOR := java -jar ~/openapi-generator-cli.jar
CONFIG_FILE := ./config.yaml

migrate-up:
	KV_VIPER_FILE=$(CONFIG_FILE) go build -o ./cmd/voting-svc/main ./cmd/voting-svc/main.go
	KV_VIPER_FILE=$(CONFIG_FILE) ./cmd/voting-svc/main migrate up

migrate-down:
	KV_VIPER_FILE=$(CONFIG_FILE) go build -o ./cmd/voting-svc/main ./cmd/voting-svc/main.go
	KV_VIPER_FILE=$(CONFIG_FILE) ./cmd/voting-svc/main migrate down

run-server:
	KV_VIPER_FILE=$(CONFIG_FILE) go build -o ./cmd/voting-svc/main ./cmd/voting-svc/main.go
	KV_VIPER_FILE=$(CONFIG_FILE) ./cmd/voting-svc/main run service

docker-uo:
	docker compose up -d

docker-down:
	docker compose down

docker-rebuild:
	docker compose up -d --build --force-recreate