package sonarqube

type PullRequest struct {
	Key    string `json:"key"`
	Branch string `json:"branch"`
	URL    string `json:"url"`
}

type ProjectPullRequests struct {
	PullRequests []PullRequest `json:"pullRequests"`
}
