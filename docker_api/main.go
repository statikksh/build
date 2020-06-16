package docker_api

import (
	"context"
	"io"

	DockerTypes "github.com/docker/docker/api/types"
	DockerContainer "github.com/docker/docker/api/types/container"
	DockerClient "github.com/docker/docker/client"
)

// Represents a Docker client.
type Docker struct {
	Context context.Context
	Client  *DockerClient.Client
}

// StopBuildContainer Stops a build container.
func (docker *Docker) StopBuildContainer(containerName string) (error error) {
	error = docker.Client.ContainerRemove(docker.Context, containerName, DockerTypes.ContainerRemoveOptions{
		Force: true,
	})

	return error
}

// StartBuildContainer Starts a new build container for with repository.
func (docker *Docker) StartBuildContainer(containerName string, repository string) (containerID string, error error) {
	container, error := docker.Client.ContainerCreate(docker.Context, &DockerContainer.Config{
		Image: "statikk:build",
		Env:   []string{"REPOSITORY=" + repository},
	}, nil, nil, containerName)

	error = docker.Client.ContainerStart(docker.Context, container.ID, DockerTypes.ContainerStartOptions{})
	return container.ID, error
}

// GetLogStreamFromContainer Returns the log stream from a container by the ID.
func (docker *Docker) GetLogStreamFromContainer(containerId string) (io.ReadCloser, error) {
	return docker.Client.ContainerLogs(docker.Context, containerId, DockerTypes.ContainerLogsOptions{
		Follow:     true,
		ShowStderr: true,
		ShowStdout: true,
		Timestamps: true,
	})
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
