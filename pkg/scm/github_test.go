package scm

import (
	"context"
	_ "embed"
	"io"
	"net/http"
	"testing"

	"github.com/google/go-github/v41/github"
	"github.com/herlon214/sonarqube-pr-issues/pkg/sonarqube"
	"github.com/migueleliasweb/go-github-mock/src/mock"
	"github.com/sourcegraph/go-diff/diff"

	"github.com/stretchr/testify/assert"
)

//go:embed testdata/raw_pr1.diff
var RawPrDiff string

func TestNewGithub(t *testing.T) {
	ctx := context.Background()

	gh := NewGithub(ctx, sonarqube.New("", ""), "mytoken")

	assert.NotNil(t, gh)
}

func TestParseGithubDiff(t *testing.T) {
	fileDiffs, err := diff.ParseMultiFileDiff([]byte(RawPrDiff))
	assert.NoError(t, err)
	assert.Equal(t, 3, len(fileDiffs))
}

func TestGithubPublishIssuesReviewWrongSonarDiffLine(t *testing.T) {
	ctx := context.Background()

	mockedHTTPClient := mock.NewMockedHTTPClient(
		mock.WithRequestMatchHandler(
			mock.GetReposPullsByOwnerByRepoByPullNumber,
			http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.Write([]byte(RawPrDiff))
			}),
		),
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

	err := gh.PublishIssuesReviewFor(ctx, issues, pr, true)
	assert.Equal(t, "failed to find relevant issues", err.Error())
}

func TestGithubPublishIssuesReviewCorrectSonarDiffLine(t *testing.T) {
	ctx := context.Background()

	mockedHTTPClient := mock.NewMockedHTTPClient(
		mock.WithRequestMatchHandler(
			mock.GetReposPullsByOwnerByRepoByPullNumber,
			http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.Write([]byte(RawPrDiff))
			}),
		),
		mock.WithRequestMatchHandler(
			mock.PostReposPullsReviewsByOwnerByRepoByPullNumber,
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				body, err := io.ReadAll(r.Body)
				assert.NoError(t, err)

				assert.Equal(t, `{"body":":wave: Hey, I added 1 comments about your changes, please take a look :slightly_smiling_face:","event":"REQUEST_CHANGES","comments":[{"path":"pkg/scm/github.go","body":":bug::bangbang: CRITICAL: My message ([go:S1234](root/coding_rules?open=go:S1234&rule_key=go:S1234))","side":"RIGHT","line":61}]}
`, string(body))
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
			Component: "myproject:pkg/scm/github.go",
			Severity:  "CRITICAL",
			Type:      "BUG",
			Rule:      "go:S1234",
			Message:   "My message",
			Line:      61,
		},
	}

	err := gh.PublishIssuesReviewFor(ctx, issues, pr, true)
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
