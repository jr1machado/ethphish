// Package queue provides a thin RabbitMQ client used to dispatch outbound
// email sends outside of the in-process channel the worker previously used
// directly. It exists to make an in-flight send durable across a server
// crash and to give unexpected processing failures (as opposed to modeled
// SMTP errors, which models.MailLog already retries via exponential
// backoff) a bounded, visible retry path instead of silently vanishing.
package queue

import (
	"context"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	// MailSendQueue holds one message per MailLog ready to be sent.
	MailSendQueue = "mail.send"
	// MailRetryQueue delays a message before it is dead-lettered back onto
	// MailSendQueue. The delay is implemented with a per-queue TTL plus a
	// dead-letter-exchange pointing back at the default exchange, so no
	// RabbitMQ plugin is required.
	MailRetryQueue = "mail.send.retry"
	// MailDeadQueue is the terminal home for messages that failed to
	// process (Go-level failure, not a modeled SMTP error) more than
	// MaxProcessingRetries times. Nothing consumes it automatically; it
	// exists for operator triage.
	MailDeadQueue = "mail.send.dead"

	// RetryCountHeader tracks how many times a message has been through
	// MailRetryQueue.
	RetryCountHeader = "x-retry-count"
	// MaxProcessingRetries bounds how many times a processing failure is
	// retried before the message is routed to MailDeadQueue.
	MaxProcessingRetries = 5
	// RetryDelay is how long a message waits on MailRetryQueue before it
	// is dead-lettered back onto MailSendQueue for another attempt.
	RetryDelay = 30 * time.Second
)

// Client wraps a RabbitMQ connection and channel with the topology this
// package depends on already declared.
type Client struct {
	conn *amqp.Connection
	ch   *amqp.Channel
}

// Connect dials url, opens a channel, and idempotently declares the durable
// queues used for dispatch, delayed retry, and dead-lettering.
func Connect(url string) (*Client, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, err
	}
	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, err
	}
	if err := declareTopology(ch); err != nil {
		ch.Close()
		conn.Close()
		return nil, err
	}
	return &Client{conn: conn, ch: ch}, nil
}

func declareTopology(ch *amqp.Channel) error {
	if _, err := ch.QueueDeclare(MailDeadQueue, true, false, false, false, nil); err != nil {
		return err
	}
	if _, err := ch.QueueDeclare(MailSendQueue, true, false, false, false, nil); err != nil {
		return err
	}
	_, err := ch.QueueDeclare(MailRetryQueue, true, false, false, false, amqp.Table{
		"x-message-ttl":             int64(RetryDelay / time.Millisecond),
		"x-dead-letter-exchange":    "",
		"x-dead-letter-routing-key": MailSendQueue,
	})
	return err
}

// Close releases the underlying channel and connection.
func (c *Client) Close() error {
	c.ch.Close()
	return c.conn.Close()
}

// Publish sends body to queue as a persistent message carrying retryCount in
// its headers.
func (c *Client) Publish(ctx context.Context, queue string, body []byte, retryCount int) error {
	return c.ch.PublishWithContext(ctx, "", queue, false, false, amqp.Publishing{
		ContentType:  "application/octet-stream",
		DeliveryMode: amqp.Persistent,
		Headers:      amqp.Table{RetryCountHeader: retryCount},
		Body:         body,
	})
}

// retryCountFromHeaders reads RetryCountHeader off a delivery, defaulting to
// zero for messages published without it.
func retryCountFromHeaders(headers amqp.Table) int {
	v, ok := headers[RetryCountHeader]
	if !ok {
		return 0
	}
	switch n := v.(type) {
	case int32:
		return int(n)
	case int64:
		return int(n)
	case int:
		return n
	default:
		return 0
	}
}

// nextRetryTarget decides where a failed delivery should go next: another
// trip through the delay queue, or the dead queue once retries are
// exhausted. It's a pure function so the retry/DLQ boundary can be tested
// without a broker.
func nextRetryTarget(retryCount int) (queue string, newRetryCount int) {
	if retryCount >= MaxProcessingRetries {
		return MailDeadQueue, retryCount
	}
	return MailRetryQueue, retryCount + 1
}

// Consume starts concurrency goroutines pulling deliveries off MailSendQueue
// and passing each delivery's body to handler.
//
//   - handler returning nil acks the delivery. This covers every modeled
//     outcome (sent, permanently failed, backed off for a later attempt)
//     since models.MailLog.Backoff/Error/Success already persisted the
//     outcome before handler returned.
//   - handler returning an error means processing itself failed (e.g. the
//     database was unreachable) rather than the send. The delivery is
//     republished onto the retry queue with an incremented count, or onto
//     the dead queue once MaxProcessingRetries is exceeded, and the
//     original delivery is acked either way so it isn't redelivered twice.
//
// Consume blocks until ctx is done.
func (c *Client) Consume(ctx context.Context, concurrency int, handler func(ctx context.Context, body []byte) error) error {
	deliveries, err := c.ch.Consume(MailSendQueue, "", false, false, false, false, nil)
	if err != nil {
		return err
	}
	done := make(chan struct{}, concurrency)
	for i := 0; i < concurrency; i++ {
		go func() {
			defer func() { done <- struct{}{} }()
			for {
				select {
				case <-ctx.Done():
					return
				case d, ok := <-deliveries:
					if !ok {
						return
					}
					c.handleDelivery(ctx, d, handler)
				}
			}
		}()
	}
	<-ctx.Done()
	for i := 0; i < concurrency; i++ {
		<-done
	}
	return nil
}

func (c *Client) handleDelivery(ctx context.Context, d amqp.Delivery, handler func(ctx context.Context, body []byte) error) {
	if err := handler(ctx, d.Body); err != nil {
		retryCount := retryCountFromHeaders(d.Headers)
		target, nextCount := nextRetryTarget(retryCount)
		if pubErr := c.Publish(ctx, target, d.Body, nextCount); pubErr != nil {
			// The broker connection itself is unhealthy; nack with requeue
			// so the message isn't lost, and let the caller's reconnect
			// loop recover the connection.
			d.Nack(false, true)
			return
		}
	}
	d.Ack(false)
}
