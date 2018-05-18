package api

// Job status
const (
	JobStatusNew    = "new"
	JobStatusQueued = "queued"
	JobStatusDone   = "done"
)

// Job queue names
const (
	QueueNameJobs       = "puppet-master-jobs"
	QueueNameJobResults = "puppet-master-job-results"
)

// Logger field names
const (
	LogFieldJobID = "job_id"
)

// HTTP header constants
const (
	ContentTypeHeader = "Content-Type"
	ContentTypeJSON   = "application/json"
)
