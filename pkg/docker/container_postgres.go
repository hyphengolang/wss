package docker

import (
	"context"
	"fmt"
	"time"

	"github.com/docker/go-connections/nat"
	"github.com/hyphengolang/socialize/internal/docker/options"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// PostgresContainer represents the postgres container type used in the module
type PostgresContainer struct {
	testcontainers.Container
}

// setupPostgres creates an instance of the postgres container type
func NewPostgresContainer(ctx context.Context, opts ...options.ContainerOption) (*PostgresContainer, error) {
	req := testcontainers.ContainerRequest{
		Image:        "postgres:14-alpine",
		Env:          map[string]string{},
		ExposedPorts: []string{},
		Cmd:          []string{"postgres", "-c", "fsync=off"},
	}

	for _, opt := range opts {
		opt(&req)
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, err
	}

	return &PostgresContainer{Container: container}, nil
}

func NewPostgresConnection(ctx context.Context, natPort string, timeout time.Duration, migrationSource string) (*PostgresContainer, *pgxpool.Pool, error) {
	var (
		user   = "postgres"
		pass   = "postgres"
		dbName = "test-db"
	)

	container, err := NewPostgresContainer(
		ctx,
		options.WithPort(natPort),
		options.WithInitialDatabase(user, pass, dbName),
		options.WithWaitStrategy(wait.ForLog("database system is ready to accept connections").WithOccurrence(2).WithStartupTimeout(timeout)))
	if err != nil {
		return nil, nil, err
	}

	port, err := container.MappedPort(ctx, nat.Port(natPort))
	if err != nil {
		container.Terminate(ctx)
		return nil, nil, err
	}

	host, err := container.Host(ctx)
	if err != nil {
		container.Terminate(ctx)
		return nil, nil, err
	}

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", host, port.Port(), user, pass, dbName)
	pool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		container.Terminate(ctx)
		return nil, nil, err
	}

	// make migration here
	_, err = pool.Exec(ctx, migrationSource)
	return container, pool, err
}
