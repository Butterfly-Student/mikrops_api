package xendit

import (
	"os"

	"go-template/utils/log"
)

var (
	XenditClient *Client
	IsEnabled    bool
)

func InitClient() {
	secretKey := os.Getenv("XENDIT_SECRET_KEY")
	if secretKey == "" {
		log.WithContext(nil).Warn("Xendit client not initialized: XENDIT_SECRET_KEY not set")
		IsEnabled = false
		return
	}

	env := Environment(os.Getenv("XENDIT_ENVIRONMENT"))
	if env == "" {
		env = EnvironmentDevelopment
	}

	XenditClient = NewClient(secretKey, env)

	webhookToken := os.Getenv("XENDIT_WEBHOOK_TOKEN")
	if webhookToken != "" {
		XenditClient.SetWebhookToken(webhookToken)
	}

	IsEnabled = true
	log.WithContext(nil).Info("Xendit client initialized")
}

func GetClient() *Client {
	return XenditClient
}

func IsClientEnabled() bool {
	return IsEnabled
}
