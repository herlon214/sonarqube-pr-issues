package scm

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParsePullRequestUrl(t *testing.T) {
	ghPath, err := parseGithubPath("https://github.com/herlon214/sonarqube-pr-issues/pull/2")
	assert.NoError(t, err)

	assert.Equal(t, "herlon214", ghPath.Owner)
	assert.Equal(t, "sonarqube-pr-issues", ghPath.Repo)

}
