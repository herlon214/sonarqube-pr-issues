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
func (i Issue) MarkdownMessage(root string) string {
	return fmt.Sprintf(`%s%s %s: %s ([%s](%s))`, i.TypeEmoji(), i.SeverityEmoji(), i.Severity, i.Message, i.Rule, i.RuleLink(root))
}

// RuleLink creates the url to the given rule
func (i Issue) RuleLink(root string) string {
	return fmt.Sprintf("%s/coding_rules?open=%s&rule_key=%s", root, i.Rule, i.Rule)
}

// FilePath returns the file path by reading the component and removing the project from it
func (i Issue) FilePath() string {
	return strings.Replace(i.Component, fmt.Sprintf("%s:", i.Project), "", -1)
}

// SeverityEmoji creates a nice emoji for the current severity
func (i Issue) SeverityEmoji() string {
	switch i.Severity {
	case "CRITICAL":
		return ":bangbang:"
	default:
		return ""

	}
}

// TypeEmoji creates a nice emoji for the current type
func (i Issue) TypeEmoji() string {
	switch i.Type {
	case "BUG":
		return ":bug:"
	case "CODE_SMELL":
		return ":biohazard:"
	case "VULNERABILITY":
		return ":key:"
	default:
		return ":thought_balloon:"
	}
}
