.PHONY: dev,cli

dev:
	go run .

cli:
	sqlite3 wss.db

# migration:
# 	.read ./sqlite/migrations/up.sql