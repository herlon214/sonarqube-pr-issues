package sonarqube

type WebhookInput struct {
	Status  string `json:"status"`
	Project struct {
		Key string `json:"key"`
	}
}
