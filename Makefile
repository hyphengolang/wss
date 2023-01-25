.PHONY: cli

cli:
	sqlite3 wss.db

# migration:
# 	.read ./sqlite/migrations/up.sql