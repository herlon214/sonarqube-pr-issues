package sonarqube

type Issues struct {
	Issues []Issue `json:"issues"`
}

// FilterByStatus filters the issues by the given status
func (i Issues) FilterByStatus(status string) Issues {
	filtered := make([]Issue, 0)
	for _, issue := range i.Issues {
		if issue.Status == status {
			filtered = append(filtered, issue)
		}
	}

	return Issues{Issues: filtered}
}

// FilterOutByTag filters out the issues that contains the given tag
func (i Issues) FilterOutByTag(tag string) Issues {
	filtered := make([]Issue, 0)

	for _, issue := range i.Issues {
		isTagFound := false
		for _, issueTag := range issue.Tags {
			if issueTag == tag {
				isTagFound = true
			}
		}

		if !isTagFound {
			filtered = append(filtered, issue)
		}
	}

	return Issues{Issues: filtered}
}
