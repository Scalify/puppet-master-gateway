package gateway

import "gitlab.com/scalifyme/puppet-master/puppet-master/pkg/api"

type db interface {
	Get(id string) (*api.Job, error)
	Save(job *api.Job) error
	Delete(job *api.Job) error
}
