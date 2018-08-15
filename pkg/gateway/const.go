package gateway

const (
	jsonErrFailedToDecodeBody = "{\"error\":\"Failed to decode json body\", \"message\": %q}"
	jsonErrFailedToFetchJobs   = "{\"error\":\"Failed to fetch job list\", \"message\": %q}"
	jsonErrFailedToFetchJob   = "{\"error\":\"Failed to fetch job\", \"message\": %q}"
	jsonErrFailedToSaveJob    = "{\"error\":\"Failed to save job\", \"message\": %q}"
	jsonErrFailedToDeleteJob  = "{\"error\":\"Failed to delete job\", \"message\": %q}"
	jsonErrJobNotFound        = "{\"error\":\"Job %s not found\", \"message\": %q}"

	task                  = "task"
	taskConsumeJobResults = "consumeJobResults"
	taskProduceJobs       = "produceJobs"
)
