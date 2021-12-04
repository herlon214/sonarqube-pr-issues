package sonarqube

import (
	"fmt"
	"strings"
)

type Issue struct {
	Severity  string   `json:"severity"`
	Component string   `json:"component"`
	Project   string   `json:"project"`
	Status    string   `json:"status"`
	Rule      string   `json:"rule"`
	Key       string   `json:"key"`
	Type      string   `json:"type"`
	Tags      []string `json:"tags"`
	Line      int      `json:"line"`
	Message   string   `json:"message"`
	TextRange struct {
		StartLine   int `json:"startLine"`
		EndLine     int `json:"endLine"`
		StartOffset int `json:"startOffset"`
		EndOffset   int `json:"endOffset"`
	}
}

// MarkdownMessage creates a nice markdown message for the issue
func (i Issue) MarkdownMessage() string {
	return ""
}

// RuleLink creates the url to the given rule
func (i Issue) RuleLink(root string) string {
	return fmt.Sprintf("%s/coding_rules?open=%s&rule_key=%s", root, i.Rule, i.Rule)
}

// FilePath returns the file path by reading the component and removing the project from it
func (i Issue) FilePath() string {
	return strings.Replace(i.Component, fmt.Sprintf("%s:", i.Project), "", -1)
}
