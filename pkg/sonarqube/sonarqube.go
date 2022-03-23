package sonarqube

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/pkg/errors"
)

const (
	TAG_PUBLISHED            = "published"
	BRANCH_TYPE_PULL_REQUEST = "PULL_REQUEST"
)

type BulkActionResponse struct {
	Total    int `json:"total"`
	Success  int `json:"success"`
	Ignored  int `json:"ignored"`
	Failures int `json:"failures"`
}

type Sonarqube struct {
	Root   string
	ApiKey string

	httpClient *http.Client
}

// New creates a new Sonarqube instance
func New(root string, apiKey string) *Sonarqube {
	// Create a new http client with timeout
	httpClient := &http.Client{
		Timeout: time.Second * 10,
	}

	return &Sonarqube{
		Root:   root,
		ApiKey: apiKey,

		httpClient: httpClient,
	}
}

// ProjectPullRequests reads all the PRs for the given project ID
func (s *Sonarqube) ProjectPullRequests(projectId string) (*ProjectPullRequests, error) {
	// Create a new request
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/project_pull_requests/list?project=%s", s.Root, projectId), nil)
	if err != nil {
		return nil, err
	}

	// Auth
	req.SetBasicAuth(s.ApiKey, "")

	// Execute request
	res, err := s.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	// Read response body
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	// Parse response body
	var data ProjectPullRequests
	err = json.Unmarshal(body, &data)
	if err != nil {
		return nil, err
	}

	return &data, nil
}

// FindPRForKey searches the pull request for the given project and key
func (s *Sonarqube) FindPRForKey(project string, key string) (*PullRequest, error) {
	// Fetch project pull requests
	pullRequests, err := s.ProjectPullRequests(project)
	if err != nil {
		return nil, err
	}

	// Filter by key
	for _, item := range pullRequests.PullRequests {
		if item.Key == key {
			return &item, nil
		}
	}

	return nil, errors.New("not found")
}

// FindPRForBranch searches the pull request for the given project and branch
func (s *Sonarqube) FindPRForBranch(project string, branch string) (*PullRequest, error) {
	// Fetch project pull requests
	pullRequests, err := s.ProjectPullRequests(project)
	if err != nil {
		return nil, err
	}

	// Filter by branch
	for _, item := range pullRequests.PullRequests {
		if item.Branch == branch {
			return &item, nil
		}
	}

	return nil, errors.New("not found")
}

// ListIssuesForPR list the issues for the given project and PR
func (s *Sonarqube) ListIssuesForPR(project string, prNumber string) (*Issues, error) {
	// Create a new request
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/issues/search?pullRequest=%s&componentKeys=%s", s.Root, prNumber, project), nil)
	if err != nil {
		return nil, err
	}

	// Auth
	req.SetBasicAuth(s.ApiKey, "")

	// Execute request
	res, err := s.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	// Read body
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	// Parse body
	var data Issues
	err = json.Unmarshal(body, &data)
	if err != nil {
		return nil, err
	}

	return &data, nil
}

// TagIssues adds a given tag into the given issues
func (s *Sonarqube) TagIssues(issues []Issue, tags string) (*BulkActionResponse, error) {
	issueKeys := make([]string, 0)
	for _, issue := range issues {
		issueKeys = append(issueKeys, issue.Key)
	}

	// Request URL
	params := fmt.Sprintf("issues=%s&add_tags=%s&do_transition=setinreview", strings.Join(issueKeys, ","), tags)

	// Create a new request
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/api/issues/bulk_change?%s", s.Root, params), nil)
	if err != nil {
		return nil, err
	}

	// Auth
	req.SetBasicAuth(s.ApiKey, "")

	// Execute request
	res, err := s.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	// Read body
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read tag issues response")
	}

	var bulkRes BulkActionResponse
	err = json.Unmarshal(body, &bulkRes)
	if err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal tag issues response")
	}

	return &bulkRes, nil
}
