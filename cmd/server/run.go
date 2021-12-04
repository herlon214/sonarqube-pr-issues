package server

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/herlon214/sonarqube-pr-issues/scm"
	"github.com/herlon214/sonarqube-pr-issues/sonarqube"
)

var RunCmd = &cobra.Command{
	Use:   "run",
	Short: "Starts the webhook server",
	Run:   Run,
}

func Run(cmd *cobra.Command, args []string) {
	// Context
	ctx := context.Background()

	// Environment
	port := os.Getenv("PORT")
	if port == "" {
		logrus.Panicln("PORT is required")

		return
	}
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
	ghToken := os.Getenv("GH_TOKEN")
	if ghToken == "" {
		logrus.Panicln("GH_TOKEN environment variable is missing")

		return
	}
	webhookSecret := os.Getenv("WEBHOOK_SECRET")
	if webhookSecret == "" {
		logrus.Panicln("WEBHOOK_SECRET environment variable is missing")

		return
	}

	// Sonarqube
	sonar := sonarqube.New(sonarRootURL, apiKey)
	var gh scm.SCM = scm.NewGithub(ctx, sonar, ghToken)

	// Listen
	http.HandleFunc("/webhook", WebhookHandler(webhookSecret, sonar, gh))

	logrus.Infoln("Listening on port", port)
	if err := http.ListenAndServe(fmt.Sprintf(":%s", port), nil); err != nil {
		panic(err)
	}
}

func WebhookHandler(webhookSecret string, sonar *sonarqube.Sonarqube, gh scm.SCM) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		// Read webhook secret
		reqSecret := req.Header.Get("X-Sonar-Webhook-HMAC-SHA256")
		if reqSecret == "" {
			w.WriteHeader(http.StatusUnauthorized)

			return
		}

		// Read request body
		body, err := io.ReadAll(req.Body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)

			return
		}

		// Generate hmac hash
		h := hmac.New(sha256.New, []byte(webhookSecret))

		// Write Data to it
		h.Write(body)

		// Get result and encode as hexadecimal string
		sha := hex.EncodeToString(h.Sum(nil))

		// Compare hashes
		if sha != reqSecret {
			w.WriteHeader(http.StatusUnauthorized)

			return
		}

		// Unmarshal data
		var webhook sonarqube.WebhookData
		err = json.Unmarshal(body, &webhook)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)

			return
		}

		logrus.Infoln("Processing", webhook.Project.Key, "->", webhook.Branch.Name)

		// Process the event
		err = PublishIssues(req.Context(), sonar, gh, webhook.Project.Key, webhook.Branch.Name)
		if err != nil {
			logrus.WithError(err).Errorln("Failed to publish issues for", webhook.Project.Key, webhook.Branch.Name)

			return
		}

		logrus.Infoln("Issues published for", webhook.Project.Key, webhook.Branch.Name)
	}
}

// PublishIssues publishes the issues in the PR for the given project branch
func PublishIssues(ctx context.Context, sonar *sonarqube.Sonarqube, projectScm scm.SCM, project string, branch string) error {
	// Find PR
	pr, err := sonar.FindPRForBranch(project, branch)
	if err != nil {
		return errors.Wrap(err, "failed to find PR for the given branch")
	}

	// List issues
	issues, err := sonar.ListIssuesForPR(project, pr.Key)
	if err != nil {
		return errors.Wrap(err, "failed to list issues for the given PR")
	}

	// Filter issues
	issues = issues.FilterByStatus("OPEN").FilterOutByTag(sonarqube.TAG_PUBLISHED)

	// No issues found
	if len(issues.Issues) == 0 {
		return nil
	}

	// Publish review
	err = projectScm.PublishIssuesReviewFor(ctx, issues.Issues, pr)
	if err != nil {
		return errors.Wrap(err, "Failed to publish issues review")
	}

	// Tag published issues
	bulkActionRes, err := sonar.TagIssues(issues.Issues, sonarqube.TAG_PUBLISHED)
	if err != nil {
		return errors.Wrap(err, "failed to mark issues as published")
	}

	logrus.Infoln("--------------------------")
	logrus.Infoln("Mark as published result:")
	logrus.Infoln(bulkActionRes.Success, "issues marked")
	logrus.Infoln(bulkActionRes.Ignored, "issues ignored")
	logrus.Infoln(bulkActionRes.Failures, "issues failed")
	logrus.Infoln("--------------------------")

	return nil

}
