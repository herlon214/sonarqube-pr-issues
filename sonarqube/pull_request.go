package sonarqube

type PullRequest struct {
	Key    string `json:"key"`
	Branch string `json:"branch"`
}

type ProjectPullRequests struct {
	PullRequests []PullRequest `json:"pullRequests"`
}
