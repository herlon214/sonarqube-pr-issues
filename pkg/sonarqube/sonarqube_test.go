package sonarqube

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewSonarqube(t *testing.T) {
	sonar := New("https://my-root", "apiKey")

	assert.NotNil(t, sonar)
	assert.Equal(t, "https://my-root", sonar.Root)
	assert.Equal(t, "apiKey", sonar.ApiKey)
}

func TestSonarqubeProjectPRs(t *testing.T) {
	// Mock response
	expected := `{"pullRequests":[{"key":"3","title":"Feat/newtest","branch":"feat/newtest","base":"feat/mvp","status":{"qualityGateStatus":"ERROR","bugs":2,"vulnerabilities":0,"codeSmells":0},"analysisDate":"2021-12-04T15:44:18+0000","url":"https://github.com/myorg/myproject/pull/3","target":"feat/mvp"},{"key":"2","title":"test PR","branch":"feat/test","base":"feat/mvp","status":{"qualityGateStatus":"ERROR","bugs":1,"vulnerabilities":0,"codeSmells":2},"analysisDate":"2021-12-03T18:10:59+0000","url":"https://github.com/myorg/myproject/pull/2","target":"feat/mvp"}]}`
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(expected))
	}))
	defer svr.Close()

	// New sonar
	sonar := New(svr.URL, "myapikey")

	// Read PRs
	prs, err := sonar.ProjectPullRequests("myproject")
	assert.NoError(t, err)

	assert.Equal(t, 2, len(prs.PullRequests))

	assert.Equal(t, "feat/newtest", prs.PullRequests[0].Branch)
	assert.Equal(t, "3", prs.PullRequests[0].Key)
	assert.Equal(t, "https://github.com/myorg/myproject/pull/3", prs.PullRequests[0].URL)

	assert.Equal(t, "feat/test", prs.PullRequests[1].Branch)
	assert.Equal(t, "2", prs.PullRequests[1].Key)
	assert.Equal(t, "https://github.com/myorg/myproject/pull/2", prs.PullRequests[1].URL)
}

func TestSonarqubeFindPRForKey(t *testing.T) {
	// Mock response
	expected := `{"pullRequests":[{"key":"3","title":"Feat/newtest","branch":"feat/newtest","base":"feat/mvp","status":{"qualityGateStatus":"ERROR","bugs":2,"vulnerabilities":0,"codeSmells":0},"analysisDate":"2021-12-04T15:44:18+0000","url":"https://github.com/myorg/myproject/pull/3","target":"feat/mvp"},{"key":"2","title":"test PR","branch":"feat/test","base":"feat/mvp","status":{"qualityGateStatus":"ERROR","bugs":1,"vulnerabilities":0,"codeSmells":2},"analysisDate":"2021-12-03T18:10:59+0000","url":"https://github.com/myorg/myproject/pull/2","target":"feat/mvp"}]}`
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(expected))
	}))
	defer svr.Close()

	// New sonar
	sonar := New(svr.URL, "myapikey")

	// Read PRs
	pr, err := sonar.FindPRForKey("myproject", "2")
	assert.NoError(t, err)

	assert.NotNil(t, pr)
	assert.Equal(t, "feat/test", pr.Branch)
	assert.Equal(t, "2", pr.Key)
	assert.Equal(t, "https://github.com/myorg/myproject/pull/2", pr.URL)
}

func TestSonarqubeFindPRForBranch(t *testing.T) {
	// Mock response
	expected := `{"pullRequests":[{"key":"3","title":"Feat/newtest","branch":"feat/newtest","base":"feat/mvp","status":{"qualityGateStatus":"ERROR","bugs":2,"vulnerabilities":0,"codeSmells":0},"analysisDate":"2021-12-04T15:44:18+0000","url":"https://github.com/myorg/myproject/pull/3","target":"feat/mvp"},{"key":"2","title":"test PR","branch":"feat/test","base":"feat/mvp","status":{"qualityGateStatus":"ERROR","bugs":1,"vulnerabilities":0,"codeSmells":2},"analysisDate":"2021-12-03T18:10:59+0000","url":"https://github.com/myorg/myproject/pull/2","target":"feat/mvp"}]}`
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(expected))
	}))
	defer svr.Close()

	// New sonar
	sonar := New(svr.URL, "myapikey")

	// Read PRs
	pr, err := sonar.FindPRForBranch("myproject", "feat/test")
	assert.NoError(t, err)

	assert.NotNil(t, pr)
	assert.Equal(t, "feat/test", pr.Branch)
	assert.Equal(t, "2", pr.Key)
	assert.Equal(t, "https://github.com/myorg/myproject/pull/2", pr.URL)
}

func TestSonarqubeListIssuesForPR(t *testing.T) {
	// Mock response
	expected := `{"total":4,"p":1,"ps":100,"paging":{"pageIndex":1,"pageSize":100,"total":4},"effortTotal":20,"issues":[{"key":"AX2GHjk1-Wk2ioy15Nrv","rule":"go:S1763","severity":"MAJOR","component":"myorg_myproject:pkg/newwrong.go","project":"myorg_myproject","line":5,"hash":"b659aeff7ad7bdc998cb5f5d433fc6df","textRange":{"startLine":5,"endLine":5,"startOffset":1,"endOffset":13},"flows":[],"status":"OPEN","message":"Refactor this piece of code to not have any dead code after this \"return\".","effort":"5min","debt":"5min","assignee":"herlon-aguiar98297","author":"herlon214@gmail.com","tags":["cwe","unused","published"],"creationDate":"2021-12-04T15:43:23+0000","updateDate":"2021-12-04T22:27:11+0000","type":"BUG","pullRequest":"3","scope":"MAIN","quickFixAvailable":false},{"key":"AX2GHjk1-Wk2ioy15Nrw","rule":"go:S1763","severity":"MAJOR","component":"myorg_myproject:pkg/newwrong.go","project":"myorg_myproject","line":12,"hash":"b659aeff7ad7bdc998cb5f5d433fc6df","textRange":{"startLine":12,"endLine":12,"startOffset":1,"endOffset":13},"flows":[],"status":"OPEN","message":"Refactor this piece of code to not have any dead code after this \"return\".","effort":"5min","debt":"5min","assignee":"herlon-aguiar98297","author":"herlon214@gmail.com","tags":["cwe","unused","published"],"creationDate":"2021-12-04T15:43:23+0000","updateDate":"2021-12-04T22:27:11+0000","type":"BUG","pullRequest":"3","scope":"MAIN","quickFixAvailable":false},{"key":"AX2Bhc31-Wk2ioy15L4-","rule":"go:S1763","severity":"MAJOR","component":"myorg_myproject:pkg/verywrong.go","project":"myorg_myproject","hash":"4b2445785755ba48ccfe71e989c7eafa","textRange":{"startLine":26,"endLine":26,"startOffset":1,"endOffset":13},"flows":[],"resolution":"FIXED","status":"CLOSED","message":"Refactor this piece of code to not have any dead code after this \"return\".","effort":"5min","debt":"5min","assignee":"herlon-aguiar98297","author":"herlon214@gmail.com","tags":["cwe","unused"],"creationDate":"2021-12-03T18:18:21+0000","updateDate":"2021-12-03T23:35:19+0000","closeDate":"2021-12-03T23:35:19+0000","type":"BUG","pullRequest":"3","scope":"MAIN","quickFixAvailable":false},{"key":"AX2Bhc31-Wk2ioy15L4_","rule":"go:S1763","severity":"MAJOR","component":"myorg_myproject:pkg/verywrong.go","project":"myorg_myproject","hash":"4b2445785755ba48ccfe71e989c7eafa","textRange":{"startLine":25,"endLine":25,"startOffset":1,"endOffset":13},"flows":[],"resolution":"FIXED","status":"CLOSED","message":"Refactor this piece of code to not have any dead code after this \"return\".","effort":"5min","debt":"5min","assignee":"herlon-aguiar98297","author":"herlon214@gmail.com","tags":["cwe","unused"],"creationDate":"2021-12-03T18:14:15+0000","updateDate":"2021-12-04T15:44:18+0000","closeDate":"2021-12-04T15:44:18+0000","type":"BUG","pullRequest":"3","scope":"MAIN","quickFixAvailable":false}],"components":[{"key":"myorg_myproject","enabled":true,"qualifier":"TRK","name":"myproject","longName":"myproject","pullRequest":"3"},{"key":"myorg_myproject:pkg/verywrong.go","enabled":false,"qualifier":"FIL","name":"verywrong.go","longName":"pkg/verywrong.go","path":"pkg/verywrong.go","pullRequest":"3"},{"key":"myorg_myproject:pkg/newwrong.go","enabled":true,"qualifier":"FIL","name":"newwrong.go","longName":"pkg/newwrong.go","path":"pkg/newwrong.go","pullRequest":"3"}],"facets":[]}`
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(expected))
	}))
	defer svr.Close()

	// New sonar
	sonar := New(svr.URL, "myapikey")

	// Read PRs
	issues, err := sonar.ListIssuesForPR("myorg_myproject", "3")
	assert.NoError(t, err)

	assert.Equal(t, 4, len(issues.Issues))
	assert.Equal(t, "myorg_myproject:pkg/newwrong.go", issues.Issues[0].Component)
	assert.Equal(t, "myorg_myproject", issues.Issues[0].Project)
	assert.Equal(t, "MAJOR", issues.Issues[0].Severity)
	assert.Equal(t, "go:S1763", issues.Issues[0].Rule)
	assert.Equal(t, "AX2GHjk1-Wk2ioy15Nrv", issues.Issues[0].Key)
	assert.Equal(t, `Refactor this piece of code to not have any dead code after this "return".`, issues.Issues[0].Message)
	assert.Equal(t, "BUG", issues.Issues[0].Type)
	assert.Equal(t, []string{"cwe", "unused", "published"}, issues.Issues[0].Tags)
	assert.Equal(t, "OPEN", issues.Issues[0].Status)
	assert.Equal(t, 5, issues.Issues[0].Line)
}

func TestSonarqubeTagIssues(t *testing.T) {
	// Mock response
	expected := `{"total":2,"success":2,"ignored":0,"failures":0}`
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(expected))
	}))
	defer svr.Close()

	// New sonar
	sonar := New(svr.URL, "myapikey")

	// Read PRs
	bulkResponse, err := sonar.TagIssues([]Issue{{Key: "AX2GHjk1-Wk2ioy15Nrv"}, {Key: "AX2GHjk1-Wk2ioy15Nrw"}}, TAG_PUBLISHED)
	assert.NoError(t, err)

	assert.Equal(t, 2, bulkResponse.Success)
	assert.Equal(t, 2, bulkResponse.Total)
	assert.Equal(t, 0, bulkResponse.Failures)
	assert.Equal(t, 0, bulkResponse.Ignored)
}
