package scm

import (
	"context"
	"net/http"
	"testing"

	"github.com/google/go-github/v41/github"
	"github.com/herlon214/sonarqube-pr-issues/pkg/sonarqube"
	"github.com/migueleliasweb/go-github-mock/src/mock"

	"github.com/stretchr/testify/assert"
)

func TestNewGithub(t *testing.T) {
	ctx := context.Background()

	gh := NewGithub(ctx, sonarqube.New("", ""), "mytoken")

	assert.NotNil(t, gh)
}

func TestGithubPublishIssuesReview(t *testing.T) {
	ctx := context.Background()

	mockedHTTPClient := mock.NewMockedHTTPClient(
		mock.WithRequestMatchHandler(
			mock.PostReposPullsReviewsByOwnerByRepoByPullNumber,
			http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			}),
		),
	)

	gh := &Github{
		sonar:  sonarqube.New("root", "key"),
		client: github.NewClient(mockedHTTPClient),
	}

	pr := &sonarqube.PullRequest{
		Key:    "3",
		Branch: "feat/newtest",
		URL:    "https://github.com/herlon214/sonarqube-pr-issues/pull/3",
	}

	issues := []sonarqube.Issue{
		{
			Project:   "myproject",
			Component: "myproject:pkg/my_file.go",
			Severity:  "CRITICAL",
			Type:      "BUG",
			Rule:      "go:S1234",
			Message:   "My message",
			Line:      10,
		},
	}

	reviewEvent := REVIEW_EVENT_REQUEST_CHANGES

	err := gh.PublishIssuesReviewFor(ctx, issues, pr, reviewEvent)
	assert.NoError(t, err)

}

func TestParsePullRequestUrl(t *testing.T) {
	ghPath, err := parseGithubPath("https://github.com/herlon214/sonarqube-pr-issues/pull/2")
	assert.NoError(t, err)

	assert.Equal(t, "herlon214", ghPath.Owner)
	assert.Equal(t, "sonarqube-pr-issues", ghPath.Repo)
}

func TestParsePathOrgOnly(t *testing.T) {
	ghPath, err := parseGithubPath("https://github.com/herlon214")
	assert.NoError(t, err)

	assert.Equal(t, "herlon214", ghPath.Owner)
	assert.Equal(t, "", ghPath.Repo)
}

func TestParsePathEmpty(t *testing.T) {
	ghPath, err := parseGithubPath("https://github.com")

	assert.Error(t, err)
	assert.Nil(t, ghPath)
}
