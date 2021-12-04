package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/herlon214/sonarqube-pr-issues/scm"
	"os"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/herlon214/sonarqube-pr-issues/sonarqube"
)

func main() {
	// Environment
	ghToken := os.Getenv("GH_TOKEN")
	apiKey := os.Getenv("SONAR_API_KEY")
	if apiKey == "" {
		logrus.Panicln("SONAR_API_KEY environment variable is missing")

		return
	}
	sonarRootURL := os.Getenv("SONAR_ROOT_URL")
	if sonarRootURL == "" {
		logrus.Panicln("SONAR_ROOT_URL environment variable is missing")

		return
	}

	// Flags
	project := flag.String("project", "my-project", "Sonarqube project name")
	branch := flag.String("branch", "my-branch", "SCM Branch name")
	publishReview := flag.Bool("publish", false, "Publish review")
	markAsPublished := flag.Bool("markaspublished", true, "Mark the issue as published to avoid sending it again")
	flag.Parse()

	if project == nil || *project == "" || branch == nil || *branch == "" {
		logrus.Panicln("Project / branch can't be empty")

		return
	}

	// Context
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	// Sonarqube
	sonar := sonarqube.New(sonarRootURL, apiKey)

	// Find PR
	pr, err := sonar.FindPRForBranch(*project, *branch)
	if err != nil {
		logrus.WithError(err).Panicln("Failed to find PR for the given branch:", *branch)

		return
	}

	// List issues
	issues, err := sonar.ListIssuesForPR(*project, pr.Key)
	if err != nil {
		logrus.WithError(err).Panicln("Failed to find PR for the given PR:", pr.Key)

		return
	}

	// Filter issues
	issues = issues.FilterByStatus("OPEN").FilterOutByTag(sonarqube.TAG_PUBLISHED)

	// Print issues
	printIssues(sonar, issues.Issues)

	// Check if should publish the review
	if *publishReview {
		// Check if token is set
		if ghToken == "" {
			logrus.Panicln("GH_TOKEN environment variable is missing")

			return
		}

		// Setup GitHub SCM
		var gh scm.SCM = scm.NewGithub(ctx, sonar, ghToken)

		// Publish review
		err = gh.PublishIssuesReviewFor(ctx, issues.Issues, pr)
		if err != nil {
			logrus.WithError(err).Panicln("Failed to publish issues review")

			return
		}

		logrus.Infoln("Issues review published!")

		// Check if should update the issues
		if *markAsPublished {
			bulkActionRes, err := sonar.TagIssues(issues.Issues, sonarqube.TAG_PUBLISHED)
			if err != nil {
				logrus.WithError(err).Panicln("Failed to mark issues as published")

				return
			}

			logrus.Infoln("--------------------------")
			logrus.Infoln("Mark as published result:")
			logrus.Infoln(bulkActionRes.Success, "issues marked")
			logrus.Infoln(bulkActionRes.Ignored, "issues ignored")
			logrus.Infoln(bulkActionRes.Failures, "issues failed")
			logrus.Infoln("--------------------------")
		}

	}

}

func printIssues(sonar *sonarqube.Sonarqube, issues []sonarqube.Issue) {
	for _, issue := range issues {
		logrus.Infof(fmt.Sprintf("[%s] %s: %s L%d:\n\t- %s\n", issue.Status, issue.Type, issue.FilePath(), issue.Line, issue.MarkdownMessage(sonar.Root)))
	}
}
