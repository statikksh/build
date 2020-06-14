package rabbitmq

import (
	DockerAPI "github.com/statikksh/build/docker_api"
	AMQP "github.com/streadway/amqp"
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

// OpenChannel Opens a new RabbitMQ channel.
func (consumer *Consumer) OpenChannel() (error error) {
	consumer.Channel, error = consumer.Connection.Channel()
	return error
}

// DeclareQueue Declares a new RabbitMQ queue.
func (consumer *Consumer) DeclareQueue(queue string, durable bool) (resultQueue AMQP.Queue, error error) {
	resultQueue, error = consumer.Channel.QueueDeclare(queue, durable, false, false, true, nil)
	return resultQueue, error
}

// Start Starts the consumer.
func (consumer *Consumer) Start(queue AMQP.Queue) (deliveries <-chan AMQP.Delivery, error error) {
	deliveries, error = consumer.Channel.Consume(queue.Name, consumer.Tag, true, false, true, true, nil)
	return deliveries, error
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
