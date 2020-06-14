package main

import (
	"log"
	"os"

	DockerAPI "github.com/statikksh/build/docker_api"
	RabbitMQ "github.com/statikksh/build/rabbitmq"
	AMQP "github.com/streadway/amqp"
)

var AMQP_CONNECTION_URL string = os.Getenv("AMQP_CONNECTION_URL")

func HandleDelivery(consumer *RabbitMQ.Consumer, deliveries <-chan AMQP.Delivery) {
	for delivery := range deliveries {
		log.Println(delivery)
	}

	log.Println("The deliveries channel has been closed.")
	log.Fatalln("The RabbitMQ connection has been closed.")

	consumer.Done <- nil
}

func main() {
	var error error

	// output environment variables
	log.Println("AMQP_CONNECTION_URL =", AMQP_CONNECTION_URL)

	docker, error := DockerAPI.Setup()
	if error != nil {
		log.Fatalln("Cannot connect to Docker host.", error)
	}

	consumer := RabbitMQ.CreateConsumer("statikk.builder", docker)
	error = consumer.Connect(AMQP_CONNECTION_URL)
	if error != nil {
		log.Fatalln("Cannot connect to the RabbitMQ server.", error)
	}

	error = consumer.OpenChannel()
	if error != nil {
		log.Fatalln("RabbitMQ consumer failed to open channel.", error)
	}

	buildsQueue, error := consumer.DeclareQueue("builds", true)
	if error != nil {
		log.Fatalln("Cannot declare queue \"builds\".", error)
	}

	deliveries, error := consumer.Start(buildsQueue)
	if error != nil {
		log.Fatalf("Cannot start RabbitMQ consumer.", error)
	}

	log.Println("Successfully connected to the RabbitMQ server.")
	HandleDelivery(consumer, deliveries)
}
