package docker_api

import (
	"context"

	DockerClient "github.com/docker/docker/client"
)

// Represents a Docker client.
type Docker struct {
	Context context.Context
	Client  *DockerClient.Client
}

// Setup Setups the Docker client.
func Setup() (docker Docker, error error) {
	if dockerClient, error := DockerClient.NewEnvClient(); error == nil {
		docker = Docker{
			Client:  dockerClient,
			Context: context.Background(),
		}
	}

	return docker, error
}
