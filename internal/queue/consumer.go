package queue

import (
	"time"
	amqp "github.com/rabbitmq/amqp091-go"
)

type Consumer struct {
	conn *amqp.Connection
	ch   *amqp.Channel
}

func NewConsumer(url string) (*Consumer, error) {
	conn, err := amqp.DialConfig(url, amqp.Config{
		Heartbeat: 10 * time.Second,
		Dial:      amqp.DefaultDial(5 * time.Second),
	})
	if err != nil {
		return nil, err
	}
	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, err
	}
	return &Consumer{conn: conn, ch: ch}, nil
}

// Consumes one message. ok is false when the queue is empty.
func (t *Consumer) Consume(queue string) (amqp.Delivery, bool, error) {
	return t.ch.Get(queue, false)
}

func (t *Consumer) Close() {
	t.ch.Close()
	t.conn.Close()
}