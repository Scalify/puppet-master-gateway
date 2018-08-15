package testing

import (
	"github.com/Scalify/puppet-master-gateway/pkg/api"
	"github.com/Scalify/puppet-master-gateway/pkg/database"
)

// TestDB is a db implementation used for testing
type TestDB struct {
	SavedJobs, DeletedJobs, Jobs []*api.Job
}

// NewTestDB returns a new TestDB instance
func NewTestDB() *TestDB {
	return &TestDB{
		Jobs:        make([]*api.Job, 0),
		SavedJobs:   make([]*api.Job, 0),
		DeletedJobs: make([]*api.Job, 0),
	}
}

// GetListByStatus returns all jobs withing the Jobs field
func (t *TestDB) GetListByStatus(status string, page, perPage int) ([]*api.Job, error) {
	return t.Jobs, nil
}

// GetList returns all jobs withing the Jobs field
func (t *TestDB) GetList(page, perPage int) ([]*api.Job, error) {
	return t.Jobs, nil
}

// Get returns the first job from the Jobs field with an equal UUID
func (t *TestDB) Get(id string) (*api.Job, error) {
	for _, j := range t.Jobs {
		if j.UUID == id {
			return j, nil
		}
	}

	return nil, database.ErrNotFound
}

// Save adds the given job to the savedJobs field
func (t *TestDB) Save(job *api.Job) error {
	t.SavedJobs = append(t.SavedJobs, job)
	return nil
}

// Delete adds the given job to the deletedJobs field
func (t *TestDB) Delete(job *api.Job) error {
	t.DeletedJobs = append(t.DeletedJobs, job)
	return nil
}

// GetUUIDs returns the UUIDs of the given jobs
func (t *TestDB) GetUUIDs(jobs []*api.Job) (ids []string) {
	for _, j := range jobs {
		ids = append(ids, j.UUID)
	}
	return
}
