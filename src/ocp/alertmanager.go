package ocp

import "time"

// Webhook represents the Alertmanager v4 webhook payload structure.
// It mirrors the JSON schema sent by Alertmanager when a notification is triggered.
type Webhook struct {
	// Receiver is the name of the receiver in Alertmanager configuration.
	Receiver string `json:"receiver"`
	// Status indicates the overall state of the alert group (e.g., "firing", "resolved").
	Status string `json:"status"`
	// Alerts contains the list of individual alerts included in this webhook event.
	Alerts []Alert `json:"alerts"`
	// GroupLabels are labels used to group alerts together.
	GroupLabels map[string]string `json:"groupLabels"`
	// CommonLabels are labels shared by all alerts in this group.
	CommonLabels map[string]string `json:"commonLabels"`
	// CommonAnnotations are annotations shared by all alerts in this group.
	CommonAnnotations map[string]string `json:"commonAnnotations"`
	// ExternalURL is the URL of the Alertmanager instance that sent this webhook.
	ExternalURL string `json:"externalURL"`
	// Version indicates the webhook payload version.
	Version string `json:"version"`
	// GroupKey is the unique identifier for this alert group.
	GroupKey string `json:"groupKey"`
	// TruncatedAlerts indicates how many alerts were truncated (if any).
	TruncatedAlerts int `json:"truncatedAlerts"`
}

// Alert represents a single alert object within the Alertmanager webhook payload.
type Alert struct {
	// Status indicates the state of this alert ("firing" or "resolved").
	Status string `json:"status"`
	// Labels contain key-value metadata identifying the alert (e.g., alertname, severity).
	Labels map[string]string `json:"labels"`
	// Annotations contain descriptive information about the alert (summary, description, etc.).
	Annotations map[string]string `json:"annotations"`
	// StartsAt indicates when the alert started firing.
	StartsAt time.Time `json:"startsAt"`
	// EndsAt indicates when the alert was resolved (if resolved).
	EndsAt time.Time `json:"endsAt"`
	// GeneratorURL is the Prometheus URL that generated this alert.
	GeneratorURL string `json:"generatorURL"`
	// Fingerprint is a unique hash identifying this alert instance.
	Fingerprint string `json:"fingerprint"`
}

// AlertName returns the value of the "alertname" label.
// If not present, it returns a placeholder string.
func (a Alert) AlertName() string {
	if v, ok := a.Labels["alertname"]; ok && v != "" {
		return v
	}
	return "(no alertname)"
}

// BestText returns the most meaningful human-readable message from annotations.
// The priority order is:
//  1. message
//  2. description
//  3. summary
//
// If none are available, an empty string is returned.
func (a Alert) BestText() string {
	if v := a.Annotations["message"]; v != "" {
		return v
	}
	if v := a.Annotations["description"]; v != "" {
		return v
	}
	if v := a.Annotations["summary"]; v != "" {
		return v
	}
	return ""
}
