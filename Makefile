.PHONY: dev,cli,solid

dev:
	go run -race ./cmd/

cli:
	sqlite3 wss.db

solid:
	npx degit solidjs/templates/ts web

# migration:
# 	.read ./sqlite/migrations/up.sql