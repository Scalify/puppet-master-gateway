package testing

import (
	"github.com/Scalify/puppet-master-gateway/pkg/api"
	"github.com/Scalify/puppet-master-gateway/pkg/database"
)

type TestDB struct {
	SavedJobs, DeletedJobs, Jobs []*api.Job
}

func NewTestDB() *TestDB {
	return &TestDB{
		Jobs:        make([]*api.Job, 0),
		SavedJobs:   make([]*api.Job, 0),
		DeletedJobs: make([]*api.Job, 0),
	}
}

func (t *TestDB) GetByStatus(status string, limit int) ([]*api.Job, error) {
	return t.Jobs, nil
}

func (t *TestDB) Get(id string) (*api.Job, error) {
	for _, j := range t.Jobs {
		if j.UUID == id {
			return j, nil
		}
	}

	return nil, database.ErrNotFound
}

func (t *TestDB) Save(job *api.Job) error {
	t.SavedJobs = append(t.SavedJobs, job)
	return nil
}

func (t *TestDB) Delete(job *api.Job) error {
	t.DeletedJobs = append(t.DeletedJobs, job)
	return nil
}

func (t *TestDB) GetUUIDs(jobs []*api.Job) (ids []string) {
	for _, j := range jobs {
		ids = append(ids, j.UUID)
	}
	return
}
