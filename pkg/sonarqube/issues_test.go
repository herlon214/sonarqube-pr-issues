package sonarqube

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFilterIssuesByStatus(t *testing.T) {
	issues := &Issues{
		Issues: []Issue{
			{
				Status: "OPEN",
			},
			{
				Status: "CLOSED",
			},
		},
	}

	issues = issues.FilterByStatus("OPEN")
	assert.Equal(t, 1, len(issues.Issues))
	assert.Equal(t, "OPEN", issues.Issues[0].Status)
}

func TestFilterOutByTag(t *testing.T) {
	issues := &Issues{
		Issues: []Issue{
			{
				Message: "first issue",
				Tags:    []string{"first", "second"},
			},
			{
				Message: "second issue",
				Tags:    []string{"first", "third"},
			},
		},
	}

	issues = issues.FilterOutByTag("third")
	assert.Equal(t, 1, len(issues.Issues))
	assert.Equal(t, "first issue", issues.Issues[0].Message)
}
