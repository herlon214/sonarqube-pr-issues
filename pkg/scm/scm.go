package scm

import (
	"context"

	"github.com/herlon214/sonarqube-pr-issues/pkg/sonarqube"
)

type SCM interface {
	PublishIssuesReviewFor(ctx context.Context, issues []sonarqube.Issue, pr *sonarqube.PullRequest) error
}
