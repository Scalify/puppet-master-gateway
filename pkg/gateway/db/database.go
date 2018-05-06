package db

import (
	"github.com/rhinoman/couchdb-go"
	"github.com/satori/go.uuid"
	"gitlab.com/scalifyme/puppet-master/gateway/pkg/gateway"
)

type JobDB struct {
	db *couchdb.Database
}

func NewJobDB(db *couchdb.Database) *JobDB {
	return &JobDB{
		db: db,
	}
}

func (db *JobDB) Get(id string) (*gateway.Job, error) {
	var job *gateway.Job
	rev, err := db.db.Read(id, job, nil)
	if err != nil {
		return nil, err
	}

	job.Rev = rev
	return job, nil
}

func (db *JobDB) Save(job *gateway.Job) (error) {
	if job.ID == "" {
		job.ID = uuid.NewV4().String()
	}

	rev, err := db.db.Save(job, job.ID, job.Rev)
	if err != nil {
		return err
	}

	job.Rev = rev
	return nil
}

func (db *JobDB) Delete(job *gateway.Job) (error) {
	_, err := db.db.Delete(job.ID, job.Rev)
	return err
}
