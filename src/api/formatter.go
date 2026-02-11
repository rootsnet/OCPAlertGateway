package api

import (
	"strings"

	"ocp-alert-gateway/ocp"
)

// FormatSynologyText builds the final text string that will be sent
// to the Synology Chat Incoming Webhook.
//
// Behavior:
//   - Each alert becomes a single line.
//   - Format: [alertname] description/message/summary
//   - If multiple alerts exist, they are joined with newline characters.
//   - If no alerts are present, a fallback message is returned.
func FormatSynologyText(w ocp.Webhook) string {
	lines := make([]string, 0, len(w.Alerts))

	for _, a := range w.Alerts {
		name := a.AlertName()

		// Select the most appropriate annotation text.
		// Priority: description > message > summary
		text := strings.TrimSpace(a.Annotations["description"])
		if text == "" {
			text = strings.TrimSpace(a.Annotations["message"])
		}
		if text == "" {
			text = strings.TrimSpace(a.Annotations["summary"])
		}
		if text == "" {
			text = "(no description/message/summary)"
		}

		// Append formatted single-line entry
		lines = append(lines, "["+name+"] "+oneLine(text))
	}

	// If no alerts exist in payload, return fallback message
	if len(lines) == 0 {
		return "[Alertmanager] (no alerts in payload)"
	}

	// Join all alert lines with newline separator
	return strings.Join(lines, "\n")
}

// oneLine converts multi-line text into a single line by:
//   - Replacing newline characters with spaces
//   - Replacing carriage returns with spaces
//   - Trimming surrounding whitespace
//
// This ensures the message renders cleanly in chat environments.
func oneLine(s string) string {
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.ReplaceAll(s, "\r", " ")
	return strings.TrimSpace(s)
}
