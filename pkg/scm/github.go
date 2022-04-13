package scm

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/google/go-github/v41/github"
	"github.com/pkg/errors"
	"github.com/sourcegraph/go-diff/diff"
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

	// Fetch PR diffs
	ghDiff, _, err := g.client.PullRequests.GetRaw(ctx, ghPath.Owner, ghPath.Repo, prNumber, github.RawOptions{github.Diff})
	if err != nil {
		return errors.Wrap(err, "failed to get raw PR")
	}

	// Parse diffs
	fileDiffs, err := diff.ParseMultiFileDiff([]byte(ghDiff))
	if err != nil {
		return errors.Wrap(err, "failed to parse diff")
	}

	diffMap := make(map[string][]*diff.Hunk)
	for i := range fileDiffs {
		fileName := fileDiffs[i].OrigName[2:]
		diffMap[fileName] = fileDiffs[i].Hunks
	}

	comments := make([]*github.DraftReviewComment, 0)

	// Create a comment for each issue
	for _, issue := range issues {
		side := "RIGHT"
		message := issue.MarkdownMessage(g.sonar.Root)
		filePath := issue.FilePath()
		lineNumber := issue.Line

		// Skip if current issue is not part of the PR diff
		hunks, ok := diffMap[filePath]
		if !ok {
			continue
		}

		for _, hunk := range hunks {
			if lineNumber < int(hunk.OrigStartLine) || lineNumber < int(hunk.NewStartLine) {
				continue
			}
			if lineNumber > int(hunk.OrigStartLine+hunk.OrigLines) || lineNumber > int(hunk.NewStartLine+hunk.NewLines) {
				continue
			}

			comment := &github.DraftReviewComment{
				Path: &filePath,
				Body: &message,
				Side: &side,
				Line: &lineNumber,
			}
			comments = append(comments, comment)
		}

	}

	if len(comments) == 0 {
		return errors.Wrap(err, "failed to find relevant issues")
	}

	body := fmt.Sprintf(`:wave: Hey, I added %d comments about your changes, please take a look :slightly_smiling_face:`, len(issues))

	reviewRequest := &github.PullRequestReviewRequest{
		Body:     &body,
		Event:    &reviewEvent,
		Comments: comments,
	}

	// Create the review
	_, _, err = g.client.PullRequests.CreateReview(ctx, ghPath.Owner, ghPath.Repo, prNumber, reviewRequest)
	if err != nil {
		return errors.Wrap(err, "failed to create review")
	}

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
