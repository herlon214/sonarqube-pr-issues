package sonarqube

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMarkdownMessage(t *testing.T) {
	issue := Issue{
		Project:   "myproject",
		Component: "myproject:pkg/my_file.go",
		Severity:  "CRITICAL",
		Type:      "BUG",
		Rule:      "go:S1234",
		Message:   "My message",
	}

	assert.Equal(t, ":bug::bangbang: CRITICAL: My message ([go:S1234](https://my-sonar/coding_rules?open=go:S1234&rule_key=go:S1234))", issue.MarkdownMessage("https://my-sonar"))
}

func TestFilePath(t *testing.T) {
	issue := Issue{Project: "myproject", Component: "myproject:pkg/file.go"}

	assert.Equal(t, "pkg/file.go", issue.FilePath())
}

func TestTypeEmojis(t *testing.T) {
	assert.Equal(t, ":bug:", Issue{Type: "BUG"}.TypeEmoji())
	assert.Equal(t, ":biohazard:", Issue{Type: "CODE_SMELL"}.TypeEmoji())
	assert.Equal(t, ":key:", Issue{Type: "VULNERABILITY"}.TypeEmoji())
	assert.Equal(t, ":thought_balloon:", Issue{Type: "SOMETHINGELSE"}.TypeEmoji())

}
