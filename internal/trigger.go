package internal

import (
	"bufio"
	"context"
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
	fmt.Printf("Outputting logs of donwstream pipeline %v\n", p.WebURL)
	fmt.Println("---")

	// sent log bytes per job id
	logBytes := make(map[int]int64)
	jobIDs, err := t.getJobIDs(ctx, p.ID)

	for {
		time.Sleep(2 * time.Second)

		for _, jobID := range jobIDs {
			if _, ok := logBytes[jobID]; !ok {
				logBytes[jobID] = 0
			}
			if bytes, err := t.outputJobLogs(ctx, jobID, logBytes[jobID]); err != nil {
				return fmt.Errorf("failed to output job log. %w", err)
			} else {
				logBytes[jobID] = logBytes[jobID] + bytes
			}
		}

		finished, err := t.isPipelineFinished(ctx, p.ID)
		if err != nil {
			return fmt.Errorf("failed to check if pipline is finished. %w", err)
		}
		if finished {
			break
		}
	}

	fmt.Println("---")

	pipelineFailed, err := t.isPipelineFailed(ctx, p.ID)
	if err != nil {
		return fmt.Errorf("failed to output pipeline logs. %w", err)
	}
	if pipelineFailed {
		return fmt.Errorf("pipleine %v failed", p.WebURL)
	}
	return nil
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

func (t *trigger) outputJobLogs(ctx context.Context, id int, skip int64) (int64, error) {
	r, _, err := t.client.Jobs.GetTraceFile(t.config.ProjectID, id, gitlab.WithContext(ctx))
	if err != nil {
		return 0, fmt.Errorf("failed to job log file. %w", err)
	}

	bufr := bufio.NewReader(r)
	if _, err := bufr.Discard(int(skip)); err != nil {
		return 0, fmt.Errorf("failed to skip %v bytes from job log. %w", skip, err)
	}

	if bytes, err := io.Copy(os.Stdout, bufr); err != nil {
		return 0, fmt.Errorf("failed to output job log. %w", err)
	} else {
		return bytes, nil
	}
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

func (t *trigger) getJobIDs(ctx context.Context, id int) ([]int, error) {
	opt := &gitlab.ListJobsOptions{}
	jobs, _, err := t.client.Jobs.ListPipelineJobs(t.config.ProjectID, id, opt, gitlab.WithContext(ctx))
	if err != nil {
		return nil, fmt.Errorf("failed to list pipeline jobs. %w", err)
	}
	jobIDs := make([]int, len(jobs))
	for i := range jobs {
		jobIDs[i] = jobs[i].ID
	}
	return jobIDs, nil
}
