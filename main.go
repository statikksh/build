package main

import (
	"io"
	"log"
	"os"

	DockerAPI "github.com/statikksh/build/docker_api"
	LogPublisherUtil "github.com/statikksh/build/logpublisher"
	RabbitMQ "github.com/statikksh/build/rabbitmq"
	AMQP "github.com/streadway/amqp"
)

const (
	buildLogsExchange = "statikk.build.logs"
)

var AMQP_CONNECTION_URL string = os.Getenv("AMQP_CONNECTION_URL")

func HandleStartBuildContainer(consumer *RabbitMQ.Consumer, repository string, repositoryID string) {
	log.Println("Starting build container for", repositoryID)
	_, error := consumer.Docker.StartBuildContainer(repositoryID, repository)

	if error != nil {
		log.Println("Failed to start build container for", repositoryID)
		log.Printf("Reason: %s\n", error)
		return
	}

	logPublisher := LogPublisherUtil.LogPublisher{
		Channel:    consumer.Channel,
		Exchange:   buildLogsExchange,
		RoutingKey: "",
		Repository: repositoryID,
	}

	logs, error := consumer.Docker.GetLogStreamFromContainer(repositoryID)
	if error != nil {
		log.Println("Failed to get logs for container", repositoryID)
		log.Printf("Reason: %s\n", error)

		logPublisher.Write([]byte("Something wrong happened, logs are not available for this build."))
		return
	}

	if _, error := io.Copy(&logPublisher, logs); error != nil {
		log.Printf("Cannot publish log output of %s: %s\n", repositoryID, error.Error())
		return
	}
}

func HandleStopBuildContainer(consumer *RabbitMQ.Consumer, repositoryID string) {
	log.Println("Stopping build container ", repositoryID)
	error := consumer.Docker.StopBuildContainer(repositoryID)

	if error != nil {
		log.Printf("Cannot remove container %s: %s\n", repositoryID, error)
	} else {
		log.Printf("Build stoped for %s.\n", repositoryID)
	}
}

func HandleDelivery(consumer *RabbitMQ.Consumer, deliveries <-chan AMQP.Delivery) {
	for delivery := range deliveries {
		action := delivery.Headers["action"].(string)

		repository := delivery.Headers["repository"]
		repositoryID := delivery.Headers["repository-id"]

		switch action {
		case "start":
			if repository != nil && repositoryID != nil {
				go HandleStartBuildContainer(consumer, repository.(string), repositoryID.(string))
			} else {
				log.Printf("[%d] Cannot stop container.\n", delivery.DeliveryTag)
				log.Println("Reason: The field `repository-id` or `repository` is missing from message headers.")
			}
		case "stop":
			if repositoryID != nil {
				go HandleStopBuildContainer(consumer, repositoryID.(string))
			} else {
				log.Printf("[%d] Cannot stop container.\n", delivery.DeliveryTag)
				log.Println("Reason: The field `repository-id` is missing from message headers.")
			}
		}
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

	exchangeError := consumer.Channel.ExchangeDeclare(buildLogsExchange, "fanout", true, false, false, true, nil)
	if exchangeError != nil {
		log.Fatalln("Cannot declare fanout exchange ", buildLogsExchange, error)
	}

	deliveries, error := consumer.Start(buildsQueue)
	if error != nil {
		log.Fatalf("Cannot start RabbitMQ consumer.", error)
	}

	log.Println("Successfully connected to the RabbitMQ server.")
	HandleDelivery(consumer, deliveries)
}
