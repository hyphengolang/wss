	CREATE SCHEMA communications;

	CREATE EXTENSION IF NOT EXISTS pgcrypto;

	CREATE TABLE IF NOT EXISTS communications.thread (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid()
	);