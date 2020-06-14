package rabbitmq

import (
	DockerAPI "github.com/statikksh/build/docker_api"
	AMQP "github.com/streadway/aqmp"
)

// Consumer Represents a RabbitMQ consumer
type Consumer struct {
	Connection *AMQP.Connection
	Channel    *AMQP.Channel
	Tag        string

	Docker DockerAPI.Docker

	Done chan error
}

// Connect Connects to a RabbitMQ server by the connection URL
func (consumer *Consumer) Connect(amqpConnectionURL string) (error error) {
	consumer.Connection, error = AMQP.Dial(amqpConnectionURL)
	return error
}

// CreateConsumer Creates a new AQMP consumer
func CreateConsumer(consumerTag string, docker DockerAPI.Docker) *Consumer {
	consumer := &Consumer{
		Connection: nil,
		Channel:    nil,
		Docker:     docker,
		Tag:        consumerTag,
	}

	return consumer
}
