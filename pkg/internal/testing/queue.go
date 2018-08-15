package testing

import "github.com/streadway/amqp"

type TestQueue struct {
	QueuesDeclared []string
	Messages       [][]byte
}

func NewTestQueue() *TestQueue {
	return &TestQueue{
		QueuesDeclared: make([]string, 0),
		Messages:       make([][]byte, 0),
	}
}

func (t *TestQueue) QueueDeclare(name string, durable, autoDelete, exclusive, noWait bool, args amqp.Table) (amqp.Queue, error) {
	return amqp.Queue{
		Name: name,
	}, nil
}

func (t *TestQueue) Consume(queue, consumer string, autoAck, exclusive, noLocal, noWait bool, args amqp.Table) (<-chan amqp.Delivery, error) {
	c := make(chan amqp.Delivery, len(t.Messages))

	for _, msg := range t.Messages {
		c <- amqp.Delivery{
			Body:         msg,
			Acknowledger: &Acknowledger{},
		}
	}

	return c, nil
}

func (t *TestQueue) Publish(exchange, key string, mandatory, immediate bool, msg amqp.Publishing) error {
	t.Messages = append(t.Messages, msg.Body)

	return nil
}

func (t *TestQueue) Qos(prefetchCount, prefetchSize int, global bool) error {
	return nil
}

type Acknowledger struct{}

func (a *Acknowledger) Ack(tag uint64, multiple bool) error {
	return nil
}
func (a *Acknowledger) Nack(tag uint64, multiple bool, requeue bool) error {
	return nil
}
func (a *Acknowledger) Reject(tag uint64, requeue bool) error {
	return nil
}
