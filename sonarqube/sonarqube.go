package sonarqube

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

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

func (s *Sonarqube) ProjectPullRequests(project string) (*ProjectPullRequests, error) {
	// Create a new request
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/project_pull_requests/list?project=%s", s.Root, project), nil)
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
