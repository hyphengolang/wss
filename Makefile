.PHONY: dev,cli,up,down,psql

dev:
	go run -race ./cmd/wss/

cli:
	sqlite3 wss.db

up:
	docker compose up -d

down:
	docker compose down

psql:
	psql -h localhost -U postgres -W