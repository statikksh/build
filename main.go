package main

import (
	"fmt"
	"io"
	"log"
	"os"

	DockerAPI "github.com/statikksh/build/docker_api"
	LogPublisherUtil "github.com/statikksh/build/logpublisher"
	RabbitMQ "github.com/statikksh/build/rabbitmq"
	AMQP "github.com/streadway/amqp"
	ErrorGroup "golang.org/x/sync/errgroup"
)

const (
	buildLogsExchange = "statikk.build.logs"
	statusExchange    = "statikk.build.status"
)

var AMQP_CONNECTION_URL string = os.Getenv("AMQP_CONNECTION_URL")

func HandleStartBuildContainer(consumer *RabbitMQ.Consumer, repository string, repositoryID string) error {
	log.Println("Starting build container for", repositoryID)
	_, error := consumer.Docker.StartBuildContainer(repositoryID, repository)

	if error != nil {
		log.Println("Failed to start build container for", repositoryID)
		log.Printf("Reason: %s\n", error)
		return error
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
		return error
	}

	if _, error := io.Copy(&logPublisher, logs); error != nil {
		log.Printf("Cannot publish log output of %s: %s\n", repositoryID, error.Error())
		return error
	}

	return nil
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

func SendBuildStatus(consumer *RabbitMQ.Consumer, repositoryID string, status string) {
	error := consumer.Channel.Publish(statusExchange, "", false, false, AMQP.Publishing{
		Headers: AMQP.Table{
			"repository": repositoryID,
			"status":     status,
		},
	})

	if error != nil {
		fmt.Println("Failed to send build status to", repositoryID)
		fmt.Printf("Reason: %s\n", error)
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
				errors, _ := ErrorGroup.WithContext(consumer.Docker.Context)
				errors.Go(func() error {
					return HandleStartBuildContainer(consumer, repository.(string), repositoryID.(string))
				})

				error := errors.Wait()
				if error != nil {
					log.Println("The build on container", repositoryID, "has failed.")
					SendBuildStatus(consumer, repositoryID.(string), "FAILED")
				} else {
					log.Println("The build on container", repositoryID, "has successfully ended.")
					SendBuildStatus(consumer, repositoryID.(string), "SUCCESS")
				}
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

	statusExchangeError := consumer.Channel.ExchangeDeclare(statusExchange, "fanout", true, false, false, true, nil)
	if statusExchangeError != nil {
		log.Fatalln("Cannot declare fanout exchange ", statusExchange, error)
	}

	deliveries, error := consumer.Start(buildsQueue)
	if error != nil {
		log.Fatalf("Cannot start RabbitMQ consumer.", error)
	}

	log.Println("Successfully connected to the RabbitMQ server.")
	HandleDelivery(consumer, deliveries)
}
