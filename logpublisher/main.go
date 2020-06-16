package logpublisher

import (
	AMQP "github.com/streadway/amqp"
)

// LogPublisher A io.Writer implementation to write data to a RabbitMQ queue.
type LogPublisher struct {
	Channel    *AMQP.Channel
	Exchange   string
	RoutingKey string
	Repository string
}

func (publisher *LogPublisher) Write(data []byte) (length int, error error) {
	error = publisher.Channel.Publish(publisher.Exchange, publisher.RoutingKey, false, false, AMQP.Publishing{
		Body: data,
		Headers: AMQP.Table{
			"repository": publisher.Repository,
		},
	})

	return len(data), error
}
