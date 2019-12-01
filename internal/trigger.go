package internal

import (
	"context"
	"errors"
	"github.com/xanzy/go-gitlab"
)

type Trigger interface {
	Run(ctx context.Context) error
}

type trigger struct {
	config Config
	client *gitlab.Client
}

func NewTrigger(config Config, client *gitlab.Client) Trigger {
	return &trigger{
		config: config,
		client: client,
	}
}

func (t *trigger) Run(ctx context.Context) error {
	return errors.New("implement me")
}
