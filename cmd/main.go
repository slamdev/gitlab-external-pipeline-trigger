package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/slamdev/gitlab-external-pipeline-trigger/internal"
	"github.com/xanzy/go-gitlab"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

const httpTimeout = 10 * time.Second

var config internal.Config
var variables internal.MapFlags

func init() {
	flag.IntVar(&config.ProjectID, "p-id", 0, "Project ID")
	flag.StringVar(&config.UserToken, "u-token", "", "User token")
	flag.StringVar(&config.PipelineToken, "p-token", "", "Pipeline token")
	flag.StringVar(&config.Ref, "ref", "master", "Ref")
	flag.StringVar(&config.GitlabURL, "url", "https://gitlab.com", "Gitlab URL")
	flag.DurationVar(&config.Timeout, "t", 30*time.Minute, "Timeout for waiting for pipeline to finish")
	flag.Var(&variables, "v", "Variables")
}

func main() {
	if err := parseConfig(); err != nil {
		fmt.Println(err)
		flag.Usage()
		os.Exit(2)
	}
	gitlabClient, err := createGitlabClient(config.GitlabURL, config.UserToken)
	if err != nil {
		log.Fatal(err)
	}
	t := internal.NewTrigger(config, gitlabClient)
	ctx := context.Background()
	if err := t.Run(ctx); err != nil {
		log.Fatal(err)
	}
}

func parseConfig() error {
	flag.Parse()
	config.Variables = variables
	required := []string{"u-token", "p-token", "p-id"}
	var err error
	flag.VisitAll(func(f *flag.Flag) {
		for _, r := range required {
			if r == f.Name && (f.Value.String() == "" || f.Value.String() == "0") {
				err = fmt.Errorf("%v is empty", f.Usage)
			}
		}
	})
	return err
}

func createGitlabClient(url, token string) (*gitlab.Client, error) {
	httpClient := &http.Client{Timeout: httpTimeout}
	if strings.HasSuffix(url, "/api/v4") {
		url = strings.TrimSuffix(url, "/api/v4")
	}
	gitlabClient := gitlab.NewClient(httpClient, token)
	if err := gitlabClient.SetBaseURL(url); err != nil {
		return nil, fmt.Errorf("failed to set %v as base url: %w", url, err)
	}
	return gitlabClient, nil
}
