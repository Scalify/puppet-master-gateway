package database

import (
	"github.com/Scalify/puppet-master-gateway/pkg/api"
	"github.com/rhinoman/couchdb-go"
	"github.com/satori/go.uuid"
)

// JobDB talks to a couchDB server and handles Job instances
type JobDB struct {
	db *couchdb.Database
}

// NewJobDB returns a new JobDB instance
func NewJobDB(db *couchdb.Database) *JobDB {
	return &JobDB{
		db: db,
	}
}

// Get fetches a job from database, identified by given UUID
func (db *JobDB) Get(id string) (*api.Job, error) {
	job := api.NewJob()
	rev, err := db.db.Read(id, job, nil)
	if err != nil {
		return nil, db.checkKnownErrors(err)
	}

	job.Rev = rev
	return job, nil
}

type jobList struct {
	Docs []*api.Job `json:"docs"`
}

func (db *JobDB) getListBy(selector map[string]interface{}, page, perPage int) ([]*api.Job, error) {
	result := &jobList{}
	query := &couchdb.FindQueryParams{
		Selector: selector,
		Limit: perPage,
		Skip:  perPage * (page - 1),
	}

	if err := db.db.Find(result, query); err != nil {
		return nil, err
	}

	return result.Docs, nil
}

// GetListByStatus returns a paginated list of jobs with the given status
func (db *JobDB) GetListByStatus(status string, page, perPage int) ([]*api.Job, error) {
	selector := map[string]interface{}{
		"status": map[string]interface{}{
			"$eq": status,
		},
	}

	return db.getListBy(selector, page, perPage)
}

// GetList returns a paginated list of jobs
func (db *JobDB) GetList(page, perPage int) ([]*api.Job, error) {
	return db.getListBy(map[string]interface{}{}, page, perPage)
}

// Save writes the job to DB
func (db *JobDB) Save(job *api.Job) error {
	if job.UUID == "" {
		job.UUID = uuid.NewV4().String()
	}

	rev, err := db.db.Save(job, job.UUID, job.Rev)
	if err != nil {
		return err
	}

	job.Rev = rev
	return nil
}

// Delete removes the job from the database
func (db *JobDB) Delete(job *api.Job) error {
	_, err := db.db.Delete(job.UUID, job.Rev)
	return db.checkKnownErrors(err)
}

func (db *JobDB) checkKnownErrors(err error) error {
	if err == nil {
		return nil
	}

	if couchErr, ok := err.(*couchdb.Error); ok {
		if couchErr.StatusCode == 404 {
			return ErrNotFound
		}
	}

	return err
}
