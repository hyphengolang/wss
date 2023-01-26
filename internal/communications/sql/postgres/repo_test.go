package repo_test

import (
	"context"
	"log"
	"testing"
	"time"

	repo "com.adoublef.wss/internal/communications/sql/postgres"
	"com.adoublef.wss/pkg/docker"
	"github.com/google/uuid"
	"github.com/hyphengolang/prelude/testing/is"

	comms "com.adoublef.wss/internal/communications"
)

var (
	testRepo      repo.ThreadRepo
	testContainer *docker.PostgresContainer

	testMigration = `
	CREATE SCHEMA communications;

	CREATE EXTENSION IF NOT EXISTS pgcrypto;

	CREATE TABLE IF NOT EXISTS communications.thread (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid()
	);
	`
)

func init() {
	container, conn, err := docker.NewPostgresConnection(context.Background(), "5432/tcp", 15*time.Second, testMigration)
	if err != nil {
		log.Fatal(err)
	}

	// initialize test repo
	testRepo = repo.NewChatRepo(conn)
	// initialize test container
	testContainer = container
}

func TestRepo(t *testing.T) {
	is := is.New(t)

	var testID uuid.UUID

	t.Run("add new chat thread", func(t *testing.T) {
		ctx := context.Background()

		var thread comms.Thread

		is.NoErr(testRepo.Create(ctx, &thread)) // failed to add chat thread to table

		testID = thread.ID
	})

	t.Run("get thread from table", func(t *testing.T) {
		ctx := context.Background()

		thread, err := testRepo.Find(ctx, testID)
		is.NoErr(err)               // failed to get thread from table
		is.Equal(thread.ID, testID) // thread id does not match
	})

	t.Run("delete thread from table", func(t *testing.T) {
		ctx := context.Background()

		is.NoErr(testRepo.Delete(ctx, testID)) // failed to delete thread from table
	})

	t.Run("get all threads from table", func(t *testing.T) {
		ctx := context.Background()

		threads, err := testRepo.FindMany(ctx)
		is.NoErr(err) // failed to get all threads from table
		is.Equal(len(threads), 0)
	})
}
