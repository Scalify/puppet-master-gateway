package cmd

type SharedEnv struct {
	Verbose bool `default:"false" split_words:"true"`
}

type CouchEnv struct {
	CouchDbHost     string `required:"true" split_words:"true"`
	CouchDbPort     int    `required:"true" split_words:"true"`
	CouchDbUsername string `required:"true" split_words:"true"`
	CouchDbPassword string `required:"true" split_words:"true"`
}

type QueueEnv struct {
	QueueHost     string `required:"true" split_words:"true"`
	QueuePort     int    `required:"true" split_words:"true"`
	QueueUsername string `required:"true" split_words:"true"`
	QueuePassword string `required:"true" split_words:"true"`
}
