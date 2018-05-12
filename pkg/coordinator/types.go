package coordinator

import (
	"github.com/streadway/amqp"
	"gitlab.com/scalifyme/puppet-master/puppet-master/pkg/api"
)

type queue interface {
	QueueDeclare(name string, durable, autoDelete, exclusive, noWait bool, args amqp.Table) (amqp.Queue, error)
	Consume(queue, consumer string, autoAck, exclusive, noLocal, noWait bool, args amqp.Table) (<-chan amqp.Delivery, error)
	Publish(exchange, key string, mandatory, immediate bool, msg amqp.Publishing) error
	Qos(prefetchCount, prefetchSize int, global bool) error
}

type db interface {
	GetByStatus(status string, limit int) ([]*api.Job, error)
	Get(id string) (*api.Job, error)
	Save(job *api.Job) error
	Delete(job *api.Job) error
}
