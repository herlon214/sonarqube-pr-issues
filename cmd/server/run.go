package server

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/avast/retry-go"
	scm2 "github.com/herlon214/sonarqube-pr-issues/pkg/scm"
	sonarqube2 "github.com/herlon214/sonarqube-pr-issues/pkg/sonarqube"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"io"
	"net/http"
	"os"
	"time"
)

var RunCmd = &cobra.Command{
	Use:              "run",
	Short:            "Starts the webhook server",
	Run:              Run,
	TraverseChildren: true,
}

func Run(cmd *cobra.Command, args []string) {
	// Context
	ctx := context.Background()

	// Environment
	if serverPort <= 0 {
		logrus.Panicln("A valid --port is required")

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
	sonar := sonarqube2.New(sonarRootURL, apiKey)
	var gh scm2.SCM = scm2.NewGithub(ctx, sonar, ghToken)

	// Listen
	http.HandleFunc("/webhook", WebhookHandler(webhookSecret, sonar, gh))

	logrus.Infoln("Listening on port", serverPort)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", serverPort), nil); err != nil {
		panic(err)
	}
}

func WebhookHandler(webhookSecret string, sonar *sonarqube2.Sonarqube, gh scm2.SCM) http.HandlerFunc {
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
		var webhook sonarqube2.WebhookData
		err = json.Unmarshal(body, &webhook)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)

			return
		}

		logrus.Infoln("Processing", webhook.Project.Key, "->", webhook.Branch.Name)

		// Process the event with retry
		err = retry.Do(func() error {
			return PublishIssues(req.Context(), sonar, gh, webhook.Project.Key, webhook.Branch.Name)
		}, retry.Delay(time.Minute), retry.Attempts(5))
		if err != nil {
			logrus.WithError(err).Errorln("Failed to publish issues for", webhook.Project.Key, webhook.Branch.Name)

			return
		}

		logrus.Infoln("Issues published for", webhook.Project.Key, webhook.Branch.Name)
	}
}

// PublishIssues publishes the issues in the PR for the given project branch
func PublishIssues(ctx context.Context, sonar *sonarqube2.Sonarqube, projectScm scm2.SCM, project string, branch string) error {
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
	issues = issues.FilterByStatus("OPEN").FilterOutByTag(sonarqube2.TAG_PUBLISHED)

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
	bulkActionRes, err := sonar.TagIssues(issues.Issues, sonarqube2.TAG_PUBLISHED)
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
