package sonarqube

type WebhookData struct {
	Status  string `json:"status"`
	Project struct {
		Key string `json:"key"`
	} `json:"project"`
	Branch struct {
		Name string `json:"name"`
		Type string `json:"type"`
	} `json:"branch"`
}
