package options

import (
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

type ContainerOption func(req *testcontainers.ContainerRequest)

func WithWaitStrategy(strategies ...wait.Strategy) ContainerOption {
	return func(req *testcontainers.ContainerRequest) {
		req.WaitingFor = wait.ForAll(strategies...).WithDeadline(1 * time.Minute)
	}
}

func WithPort(port string) ContainerOption {
	return func(req *testcontainers.ContainerRequest) {
		req.ExposedPorts = append(req.ExposedPorts, port)
	}
}

func WithInitialDatabase(user string, password string, dbName string) ContainerOption {
	return func(req *testcontainers.ContainerRequest) {
		req.Env["POSTGRES_USER"] = user
		req.Env["POSTGRES_PASSWORD"] = password
		req.Env["POSTGRES_DB"] = dbName
	}
}

var mounts = testcontainers.ContainerMount{
	Source:   nil,
	Target:   "/docker-entrypoint-initdb.d",
	ReadOnly: false,
}

func WithMounts(mounts ...testcontainers.ContainerMount) ContainerOption {
	return func(req *testcontainers.ContainerRequest) {
		if req.Mounts == nil {
			req.Mounts = []testcontainers.ContainerMount{}
		}
		req.Mounts = append(req.Mounts, mounts...)
	}
}

// var mounts = testcontainers.ContainerMount{
// 	// Source is typically either a GenericBindMountSource or a GenericVolumeMountSource
// 	// Source: "/path/to/source",
// 	// Target is the path where the mount should be mounted within the container
// 	Target: "/docker-entrypoint-initdb.d",
// 	// ReadOnly bool
// }
