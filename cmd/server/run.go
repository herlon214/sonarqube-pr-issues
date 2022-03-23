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

	// Process queue
	queue := make(chan func() error, 0)
	for i := 0; i < workers; i++ {
		go ProcessQueue(queue)
	}

	// Listen
	http.HandleFunc("/webhook", WebhookHandler(webhookSecret, sonar, gh, queue))

	logrus.Infoln("Listening on port", serverPort)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", serverPort), nil); err != nil {
		panic(err)
	}
}

func WebhookHandler(webhookSecret string, sonar *sonarqube2.Sonarqube, gh scm2.SCM, queue chan<- func() error) http.HandlerFunc {
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

		// Add event to queue
		logrus.Infoln("Adding to the queue", webhook.Project.Key, "->", webhook.Branch.Name)
		queue <- func() error {
			logrus.Infoln("Processing", webhook.Project.Key, "->", webhook.Branch.Name)

			if err := PublishIssues(context.Background(), sonar, gh, webhook.Project.Key, webhook.Branch.Name, webhook.Branch.Type); err != nil {
				return err
			}

			logrus.Infoln("Issues published for", webhook.Project.Key, webhook.Branch.Name)

			return nil
		}

		w.WriteHeader(http.StatusOK)
	}
}

// ProcessQueue is made to process the webhooks in background
// if the request takes more than 10s Sonarqube shows the message 'Server Unreachable'
func ProcessQueue(queue <-chan func() error) {
	for fn := range queue {
		err := retry.Do(fn, retry.Delay(time.Minute), retry.DelayType(retry.FixedDelay), retry.Attempts(5))
		if err != nil {
			logrus.WithError(err).Errorln("Failed to process webhook")
		}
	}
}

// PublishIssues publishes the issues in the PR for the given project branch
func PublishIssues(ctx context.Context, sonar *sonarqube2.Sonarqube, projectScm scm2.SCM, project string, branch string, branchType string) error {
	// Find PR
	var pr *sonarqube2.PullRequest
	var err error
	if branchType == sonarqube2.BRANCH_TYPE_PULL_REQUEST {
		pr, err = sonar.FindPRForKey(project, branch)
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("failed to find PR for key %s of the project %s", branch, project))
		}
	} else {
		pr, err = sonar.FindPRForBranch(project, branch)
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("failed to find PR for branch %s of the project %s", branch, project))
		}
	}

	// List issues
	issues, err := sonar.ListIssuesForPR(project, pr.Key)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("failed to list issues for the given PR branch %s of the project %s", branch, project))
	}

	// Filter issues
	issues = issues.FilterByStatus("OPEN").FilterOutByTag(sonarqube2.TAG_PUBLISHED)

	// No issues found
	if len(issues.Issues) == 0 {
		return nil
	}

	// Publish review
	err = projectScm.PublishIssuesReviewFor(ctx, issues.Issues, pr, requestChanges)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("Failed to publish issues review for branch %s of the project %s", branch, project))
	}

	// Tag published issues
	bulkActionRes, err := sonar.TagIssues(issues.Issues, sonarqube2.TAG_PUBLISHED)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("failed to mark issues as published for branch %s of the project %s", branch, project))
	}

	logrus.Infoln("--------------------------")
	logrus.Infoln("Mark as published result:")
	logrus.Infoln(bulkActionRes.Success, "issues marked")
	logrus.Infoln(bulkActionRes.Ignored, "issues ignored")
	logrus.Infoln(bulkActionRes.Failures, "issues failed")
	logrus.Infoln("--------------------------")

	return nil

}
