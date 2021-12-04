package scm

import (
	"context"

	"github.com/herlon214/sonarqube-pr-issues/sonarqube"

	"github.com/google/go-github/v41/github"
	"golang.org/x/oauth2"
)

type Github struct {
	client *github.Client
}

func NewGithub(ctx context.Context, token string) Github {
	// Token source
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)

	// Oauth2 client
	tc := oauth2.NewClient(ctx, ts)

	// Github client
	client := github.NewClient(tc)

	return Github{
		client: client,
	}
}

// PublishIssuesReviewFor publishes a review with a comment for each issue
func (g *Github) PublishIssuesReviewFor(ctx context.Context, issues []sonarqube.Issue, owner string, repo string, prNumber int) error {
	event := "REQUEST_CHANGES"
	comments := make([]*github.DraftReviewComment, 0)

	// Create a comment for each issue
	for _, issue := range issues {
		side := "RIGHT"
		message := issue.MarkdownMessage()
		comment := &github.DraftReviewComment{
			Path:      &issue.Component,
			Body:      &message,
			StartSide: &side,
			Side:      &side,
			Line:      &issue.Line,
		}
		comments = append(comments, comment)
	}

	reviewRequest := &github.PullRequestReviewRequest{
		Event:    &event,
		Comments: comments,
	}

	// Create the review
	createReviewRes, _, err := g.client.PullRequests.CreateReview(ctx, owner, repo, prNumber, reviewRequest)
	if err != nil {
		return err
	}

	// Submit
	_, _, err = g.client.PullRequests.SubmitReview(ctx, owner, repo, prNumber, *createReviewRes.ID, reviewRequest)
	if err != nil {
		return err
	}

	return nil
}
