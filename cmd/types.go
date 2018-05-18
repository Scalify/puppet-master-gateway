package cmd

// SharedEnv is a shared command environment config
type SharedEnv struct {
	Verbose bool `default:"false" split_words:"true"`
}

// CouchEnv is a couchDB command environment config
type CouchEnv struct {
	CouchDbHost     string `required:"true" split_words:"true"`
	CouchDbPort     int    `required:"true" split_words:"true"`
	CouchDbUsername string `required:"true" split_words:"true"`
	CouchDbPassword string `required:"true" split_words:"true"`
}

// QueueEnv is a queue command environment config
type QueueEnv struct {
	QueueHost     string `required:"true" split_words:"true"`
	QueuePort     int    `required:"true" split_words:"true"`
	QueueUsername string `required:"true" split_words:"true"`
	QueuePassword string `required:"true" split_words:"true"`
}
