package internal

import "time"

type Config struct {
	PipelineToken string
	UserToken     string
	ProjectID     int
	Ref           string
	GitlabURL     string
	Variables     map[string]string
	Timeout       time.Duration
}
