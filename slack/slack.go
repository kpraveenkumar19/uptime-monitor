package slack

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"encore.app/monitor"
	"encore.dev/pubsub"
)

type NotifyParams struct {
	// Text is the Slack message text to send.
	Text string `json:"text"`
}

// Notify sends a Slack message to a pre-configured channel using a
// Slack Incoming Webhook
//
//encore:api private
func Notify(ctx context.Context, p *NotifyParams) error {
	reqBody, err := json.Marshal(p)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, "POST", secrets.SlackWebhookURL, bytes.NewReader(reqBody))
	if err != nil {
		return err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("notify slack: %s: %s", resp.Status, body)
	}
	return nil
}

var secrets struct {
	// SlackWebhookURL defines the Slack webhook URL to send
	// uptime notifications to.
	SlackWebhookURL string
}

// add a Pub/Sub subscriber to automatically
// notify Slack when a monitored site goes up or down.
var _ = pubsub.NewSubscription(monitor.TransitionTopic, "slack-notification", pubsub.SubscriptionConfig[*monitor.TransitionEvent]{
	Handler: func(ctx context.Context, event *monitor.TransitionEvent) error {
		// Compose our message.
		msg := fmt.Sprintf("*%s is down!*", event.Site.URL)
		if event.Up {
			msg = fmt.Sprintf("*%s is back up.*", event.Site.URL)
		}

		// Send the Slack notification.
		return Notify(ctx, &NotifyParams{Text: msg})
	},
})
