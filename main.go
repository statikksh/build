package main

import (
	"log"

	DockerAPI "github.com/statikksh/build/docker_api"
)

func main() {
	_, error := DockerAPI.Setup()
	if error != nil {
		log.Fatalln("Cannot connect to Docker host.")
	}
}
