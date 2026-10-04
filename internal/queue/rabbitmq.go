package queue

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Publisher struct {
	url  string
	mu   sync.Mutex
	conn *amqp.Connection
	ch   *amqp.Channel
	done chan struct{}
}

func New(url string) (*Publisher, error) {
	if url == "" {
		return nil, errors.New("AMQP_URL is empty")
	}
	p := &Publisher{url: url, done: make(chan struct{})}
	if err := p.connect(); err != nil {
		return nil, err
	}
	go p.watch()
	return p, nil
}

func (p *Publisher) connect() error {
	conn, err := amqp.DialConfig(p.url, amqp.Config{
		Heartbeat: 10 * time.Second,
		Dial:      amqp.DefaultDial(5 * time.Second),
	})
	if err != nil {
		return fmt.Errorf("dial rabbitmq: %w", err)
	}
	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return fmt.Errorf("open channel: %w", err)
	}

	p.mu.Lock()
	p.conn, p.ch = conn, ch
	p.mu.Unlock()
	return nil
}

// watch reconnects with backoff when the connection drops.
func (p *Publisher) watch() {
	for {
		p.mu.Lock()
		conn := p.conn
		p.mu.Unlock()

		closed := conn.NotifyClose(make(chan *amqp.Error, 1))
		select {
		case <-p.done:
			return
		case err := <-closed:
			if err == nil {
				return
			}
			log.Printf("rabbitmq connection lost: %v", err)

			backoff := time.Second
			for {
				select {
				case <-p.done:
					return
				case <-time.After(backoff):
				}
				if err := p.connect(); err == nil {
					log.Println("rabbitmq reconnected")
					break
				}
				if backoff < 30*time.Second {
					backoff *= 2
				}
			}
		}
	}
}

// Declare creates a durable queue if it does not exist.
func (p *Publisher) Declare(name string) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	_, err := p.ch.QueueDeclare(name, true, false, false, false, nil)
	return err
}

// Publish sends a persistent message to the named queue.
func (p *Publisher) Publish(ctx context.Context, queue string, body []byte) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.ch.PublishWithContext(ctx, "", queue, false, false, amqp.Publishing{
		ContentType:  "application/json",
		DeliveryMode: amqp.Persistent,
		Timestamp:    time.Now(),
		Body:         body,
	})
}

// Health reports whether the connection is open.
func (p *Publisher) Health() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.conn == nil || p.conn.IsClosed() {
		return errors.New("rabbitmq connection closed")
	}
	return nil
}

func (p *Publisher) Close() {
	close(p.done)
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.ch != nil {
		p.ch.Close()
	}
	if p.conn != nil {
		p.conn.Close()
	}
}