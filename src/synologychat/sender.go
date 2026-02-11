package synologychat

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"ocp-alert-gateway/logger"
)

// Sender is responsible for sending formatted alert messages
// to a Synology Chat Incoming Webhook endpoint.
type Sender struct {
	// WebhookURL is the full Synology Chat Incoming Webhook URL.
	WebhookURL string
	// InsecureSkipVerify disables TLS certificate verification
	// (useful for internal/self-signed environments).
	InsecureSkipVerify bool
}

// synologyResp represents the JSON response returned by
// the Synology Chat Incoming Webhook API.
type synologyResp struct {
	Success bool `json:"success"`
	Error   *struct {
		Code   int    `json:"code"`
		Errors string `json:"errors"`
	} `json:"error"`
}

// SendText sends a single text message to Synology Chat.
//
// The message is sent using application/x-www-form-urlencoded
// with a "payload" field containing a JSON object:
//
//	payload={"text":"message body"}
//
// If debug is enabled, response status and body are logged.
// Returns an error if:
//   - webhook URL is empty
//   - request fails
//   - non-2xx HTTP status is returned
//   - Synology API returns success=false
//   - response is not valid JSON
func (s *Sender) SendText(text string, debug bool, log *logger.Logger) error {
	// Ensure webhook URL is provided
	if strings.TrimSpace(s.WebhookURL) == "" {
		return fmt.Errorf("synology webhook url is empty")
	}

	// Build JSON payload: {"text": "..."}
	payloadObj := map[string]string{"text": text}
	payloadJSON, err := json.Marshal(payloadObj)
	if err != nil {
		return fmt.Errorf("json marshal failed: %w", err)
	}

	// Encode as form data: payload=<json>
	form := url.Values{}
	form.Set("payload", string(payloadJSON))

	// Create HTTP POST request
	req, err := http.NewRequest(http.MethodPost, s.WebhookURL, strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Configure HTTP client with timeout
	client := &http.Client{Timeout: 10 * time.Second}

	// Optionally disable TLS verification
	if s.InsecureSkipVerify {
		client.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
	}

	// Send request
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Read response body (limit to 1MB for safety)
	rb, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	bodyStr := strings.TrimSpace(string(rb))

	// Debug logging of response
	if debug {
		log.Info.Printf("synology response status=%s", resp.Status)
		if bodyStr != "" {
			log.Info.Printf("synology response body=%s", bodyStr)
		}
	}

	// Check HTTP status code
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("synology non-2xx: %s body=%s", resp.Status, bodyStr)
	}

	// Parse Synology JSON response
	var sr synologyResp
	if err := json.Unmarshal(rb, &sr); err == nil {
		if !sr.Success {
			if sr.Error != nil {
				return fmt.Errorf("synology success=false code=%d errors=%s", sr.Error.Code, sr.Error.Errors)
			}
			return fmt.Errorf("synology success=false (no error details)")
		}
	} else {
		return fmt.Errorf("synology response not json: %s", bodyStr)
	}

	return nil
}
