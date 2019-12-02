package internal

import (
	"context"
	"errors"
	"fmt"
	"github.com/xanzy/go-gitlab"
	"io"
	"os"
	"time"
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
	opt := &gitlab.RunPipelineTriggerOptions{
		Ref:       gitlab.String(t.config.Ref),
		Token:     gitlab.String(t.config.PipelineToken),
		Variables: t.config.Variables,
	}
	p, _, err := t.client.PipelineTriggers.RunPipelineTrigger(t.config.ProjectID, opt, gitlab.WithContext(ctx))
	if err != nil {
		return fmt.Errorf("failed to trigger project pipeline. %w", err)
	}
	fmt.Printf("Waiting for pipline %v to finish\n", p.WebURL)
	if err := t.waitForPipelineToFinish(ctx, p.ID); err != nil {
		return fmt.Errorf("failed to wait for pipeline to finish. %w", err)
	}
	fmt.Println("Pipeline output:")
	if err := t.outputPipelineLogs(ctx, p.ID); err != nil {
		return fmt.Errorf("failed to output pipeline logs. %w", err)
	}
	pipelineFailed, err := t.isPipelineFailed(ctx, p.ID)
	if err != nil {
		return fmt.Errorf("failed to output pipeline logs. %w", err)
	}
	if pipelineFailed {
		return fmt.Errorf("pipleine %v failed", p.WebURL)
	}
	fmt.Println("Done")
	return nil
}

func (t *trigger) waitForPipelineToFinish(ctx context.Context, id int) error {
	timeout := time.After(t.config.Timeout)
	tick := time.Tick(2 * time.Second)
	for {
		select {
		case <-timeout:
			return errors.New("timed out")
		case <-tick:
			finished, err := t.isPipelineFinished(ctx, id)
			if err != nil {
				return fmt.Errorf("failed to check if pipline is finished. %w", err)
			} else if finished {
				fmt.Println()
				return nil
			} else {
				fmt.Print(".")
			}
		}
	}
}

func (t *trigger) isPipelineFinished(ctx context.Context, id int) (bool, error) {
	p, _, err := t.client.Pipelines.GetPipeline(t.config.ProjectID, id, gitlab.WithContext(ctx))
	if err != nil {
		return false, fmt.Errorf("failed to get pipeline info. %w", err)
	}
	finishedStatuses := []string{"failed", "manual", "canceled", "success", "skipped"}
	for _, status := range finishedStatuses {
		if p.Status == status {
			return true, nil
		}
	}
	return false, nil
}

func (t *trigger) outputPipelineLogs(ctx context.Context, id int) error {
	opt := &gitlab.ListJobsOptions{}
	jobs, _, err := t.client.Jobs.ListPipelineJobs(t.config.ProjectID, id, opt, gitlab.WithContext(ctx))
	if err != nil {
		return fmt.Errorf("failed to list pipeline jobs. %w", err)
	}
	for i := range jobs {
		job := jobs[i]
		fmt.Printf("Job \"%v\" (%v):\n", job.Name, job.WebURL)
		fmt.Println("---")
		if err := t.outputJobLogs(ctx, job.ID); err != nil {
			return fmt.Errorf("failed to output job logs. %w", err)
		}
	}
	return nil
}

func (t *trigger) outputJobLogs(ctx context.Context, id int) error {
	r, _, err := t.client.Jobs.GetTraceFile(t.config.ProjectID, id, gitlab.WithContext(ctx))
	if err != nil {
		return fmt.Errorf("failed to job log file. %w", err)
	}
	if _, err := io.Copy(os.Stdout, r); err != nil {
		return fmt.Errorf("failed to output job log. %w", err)
	}
	return nil
}

func (t *trigger) isPipelineFailed(ctx context.Context, id int) (bool, error) {
	p, _, err := t.client.Pipelines.GetPipeline(t.config.ProjectID, id, gitlab.WithContext(ctx))
	if err != nil {
		return false, fmt.Errorf("failed to get pipeline info. %w", err)
	}
	successStatuses := []string{"manual", "success"}
	for _, status := range successStatuses {
		if p.Status == status {
			return false, nil
		}
	}
	return true, nil
}
