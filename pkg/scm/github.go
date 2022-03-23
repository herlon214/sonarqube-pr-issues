package scm

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/google/go-github/v41/github"
	"github.com/pkg/errors"
	"golang.org/x/oauth2"

	"github.com/herlon214/sonarqube-pr-issues/pkg/sonarqube"
)

const (
	REVIEW_EVENT_COMMENT         = "COMMENT"
	REVIEW_EVENT_REQUEST_CHANGES = "REQUEST_CHANGES"
)

type Github struct {
	client *github.Client
	sonar  *sonarqube.Sonarqube
}

type GithubPath struct {
	Owner string
	Repo  string
}

func NewGithub(ctx context.Context, sonar *sonarqube.Sonarqube, token string) *Github {
	// Token source
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)

	// Oauth2 client
	tc := oauth2.NewClient(ctx, ts)

	// Github client
	client := github.NewClient(tc)

	return &Github{
		client: client,
		sonar:  sonar,
	}
}

// PublishIssuesReviewFor publishes a review with a comment for each issue
func (g *Github) PublishIssuesReviewFor(ctx context.Context, issues []sonarqube.Issue, pr *sonarqube.PullRequest, requestChanges bool) error {
	var reviewEvent string
	if requestChanges {
		reviewEvent = REVIEW_EVENT_REQUEST_CHANGES
	} else {
		reviewEvent = REVIEW_EVENT_COMMENT
	}

	comments := make([]*github.DraftReviewComment, 0)

	// Create a comment for each issue
	for _, issue := range issues {
		side := "RIGHT"
		message := issue.MarkdownMessage(g.sonar.Root)
		filePath := issue.FilePath()
		line := issue.Line

		comment := &github.DraftReviewComment{
			Path: &filePath,
			Body: &message,
			Side: &side,
			Line: &line,
		}
		comments = append(comments, comment)
	}

	body := fmt.Sprintf(`:wave: Hey, I added %d comments about your changes, please take a look :slightly_smiling_face:`, len(issues))

	reviewRequest := &github.PullRequestReviewRequest{
		Body:     &body,
		Event:    &reviewEvent,
		Comments: comments,
	}

	// Convert PR number into int
	prNumber, err := strconv.Atoi(pr.Key)
	if err != nil {
		return errors.Wrap(err, "failed to convert PR number to int")
	}

	// Parse PR path
	ghPath, err := parseGithubPath(pr.URL)
	if err != nil {
		return errors.Wrap(err, "failed to parse github path")
	}

	// Create the review
	out, res, err := g.client.PullRequests.CreateReview(ctx, ghPath.Owner, ghPath.Repo, prNumber, reviewRequest)
	if err != nil {
		return errors.Wrap(err, "failed to create review")
	}

	fmt.Println(out, res)

	return nil
}

// parseGithubPath converts the given path into GitHub path struct
func parseGithubPath(path string) (*GithubPath, error) {
	// Parse url
	parsedUrl, err := url.Parse(path)
	if err != nil {
		return nil, err
	}

	// Split directories
	dirs := strings.Split(parsedUrl.Path, "/")
	if len(dirs) <= 1 {
		return nil, errors.New("no directories specified")
	}

	if len(dirs) == 2 {
		return &GithubPath{
			Owner: dirs[1],
		}, nil
	} else {
		return &GithubPath{
			Owner: dirs[1],
			Repo:  dirs[2],
		}, nil
	}
}
