package cli

import (
	"context"
	"fmt"
	scm2 "github.com/herlon214/sonarqube-pr-issues/pkg/scm"
	sonarqube2 "github.com/herlon214/sonarqube-pr-issues/pkg/sonarqube"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"os"
	"time"
)

var RunCmd = &cobra.Command{
	Use:              "run",
	Short:            "Process the given project and branch",
	Run:              Run,
	TraverseChildren: true,
}

func Run(cmd *cobra.Command, args []string) {
	logrus.Infoln("Processing", project, "->", branch)

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

	if project == "" || branch == "" {
		logrus.Panicln("Project / branch can't be empty")

		return
	}

	// Context
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	// Sonarqube
	sonar := sonarqube2.New(sonarRootURL, apiKey)

	// Find PR
	pr, err := sonar.FindPRForBranch(project, branch)
	if err != nil {
		logrus.WithError(err).Panicln("Failed to find PR for the given branch:", branch)

		return
	}

	// List issues
	issues, err := sonar.ListIssuesForPR(project, pr.Key)
	if err != nil {
		logrus.WithError(err).Panicln("Failed to list issues for the given PR:", pr.Key)

		return
	}

	// Filter issues
	issues = issues.FilterByStatus("OPEN").FilterOutByTag(sonarqube2.TAG_PUBLISHED)
	if len(issues.Issues) == 0 {
		logrus.Infoln("No issues found!")

		return
	}

	// Print issues
	printIssues(sonar, issues.Issues)

	// Check if should publish the review
	if publishReview {
		// Check if token is set
		if ghToken == "" {
			logrus.Panicln("GH_TOKEN environment variable is missing")

			return
		}

		// Setup GitHub SCM
		var gh scm2.SCM = scm2.NewGithub(ctx, sonar, ghToken)

		// Publish review
		err = gh.PublishIssuesReviewFor(ctx, issues.Issues, pr, requestChanges)
		if err != nil {
			logrus.WithError(err).Panicln("Failed to publish issues review")

			return
		}

		logrus.Infoln("Issues review published!")
	}

	// Check if should update the issues
	if markAsPublished {
		bulkActionRes, err := sonar.TagIssues(issues.Issues, sonarqube2.TAG_PUBLISHED)
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

func printIssues(sonar *sonarqube2.Sonarqube, issues []sonarqube2.Issue) {
	for _, issue := range issues {
		logrus.Infof(fmt.Sprintf("[%s] %s: %s L%d:\n\t- %s\n", issue.Status, issue.Type, issue.FilePath(), issue.Line, issue.MarkdownMessage(sonar.Root)))
	}
}
