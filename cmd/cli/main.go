package main

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/herlon214/sonarqube-pr-issues/sonarqube"
)

func main() {
	// Keys
	apiKey := os.Getenv("SONAR_API_KEY")
	ghToken := os.Getenv("GH_TOKEN")
	sonarRootURL := os.Getenv("SONAR_ROOT_URL")

	ctx := context.Background()
	sonar := sonarqube.New(sonarRootURL, apiKey)

	item, err := sonar.FindPRForBranch("myproject", "feat/mvp")
	if err != nil {
		panic(err)
	}

	issues, err := sonar.ListIssuesForPR("myproject", item.Key)
	if err != nil {
		panic(err)
	}

	// Convert PR number into int
	prNumber, err := strconv.Atoi(item.Key)
	if err != nil {
		panic(err)
	}

}

func printIssues(issues []sonarqube.Issue) {
	for _, issue := range issues {
		fmt.Printf(fmt.Sprintf("\n[%s] %s: %s L%d:\n\t- %s", issue.Status, issue.Type, issue.Component, issue.Line, issue.Message))
	}
}
