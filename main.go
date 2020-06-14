package main

import (
	"log"
	"os"

	DockerAPI "github.com/statikksh/build/docker_api"
	RabbitMQ "github.com/statikksh/build/rabbitmq"
)

var AMQP_CONNECTION_URL string = os.Getenv("AMQP_CONNECTION_URL")

func main() {
	var error error

	docker, error := DockerAPI.Setup()
	if error != nil {
		log.Fatalln("Cannot connect to Docker host.", error)
	}

	consumer := RabbitMQ.CreateConsumer("statikk.builder", docker)
	error = consumer.Connect(AMQP_CONNECTION_URL)
	if error != nil {
		log.Fatalln("Cannot connect to the RabbitMQ server.", error)
	}

	log.Println("Successfully connected to the RabbitMQ server.")
}
