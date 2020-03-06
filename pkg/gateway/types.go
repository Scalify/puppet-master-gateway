package gateway

import (
	"github.com/scalify/puppet-master-gateway/pkg/api"
	"github.com/streadway/amqp"
)

type db interface {
	GetList(page, perPage int) ([]*api.Job, error)
	GetListByStatus(status string, page, perPage int) ([]*api.Job, error)
	Get(id string) (*api.Job, error)
	Save(job *api.Job) error
	Delete(job *api.Job) error
}

type queue interface {
	QueueDeclare(name string, durable, autoDelete, exclusive, noWait bool, args amqp.Table) (amqp.Queue, error)
	Consume(queue, consumer string, autoAck, exclusive, noLocal, noWait bool, args amqp.Table) (<-chan amqp.Delivery, error)
	Publish(exchange, key string, mandatory, immediate bool, msg amqp.Publishing) error
	Qos(prefetchCount, prefetchSize int, global bool) error
}
