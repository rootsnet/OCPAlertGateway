package webserver

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httputil"

	"ocp-alert-gateway/api"
	"ocp-alert-gateway/logger"
	"ocp-alert-gateway/ocp"
	"ocp-alert-gateway/synologychat"
)

// Server holds dependencies required to handle incoming webhooks
// and forward formatted messages to Synology Chat.
type Server struct {
	// Log provides stdout/stderr logging.
	Log *logger.Logger
	// Sender sends messages to Synology Chat. If nil, no message will be sent.
	Sender *synologychat.Sender
	// Debug enables verbose request/response logging.
	Debug bool
}

// Healthz is a simple health check endpoint.
// It returns HTTP 200 with a plain "ok" body.
func (s *Server) Healthz(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok\n"))
}

// Webhook receives Alertmanager webhook requests,
// parses the JSON payload, formats a text message, and sends it to Synology Chat.
//
// Logging behavior:
//   - When Debug is true: dumps full request (headers + body) and pretty-prints JSON,
//     and also logs the final Synology text.
//   - When Debug is false: logs a minimal "received" line (status/receiver/count),
//     plus the result of sending (success/failure).
//
// The handler always responds with HTTP 200 on successful parsing, even if
// downstream sending fails, to avoid excessive retries from Alertmanager.
func (s *Server) Webhook(w http.ResponseWriter, r *http.Request) {
	// Only POST is supported for webhook events.
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	// Read request body (limit to 10MB to prevent excessive memory usage).
	body, err := io.ReadAll(io.LimitReader(r.Body, 10<<20))
	if err != nil {
		s.Log.Err.Printf("failed to read body: %v", err)
		http.Error(w, "failed to read body", http.StatusBadRequest)
		return
	}
	_ = r.Body.Close()

	// Debug mode: dump full request (headers + body) and pretty-print JSON.
	if s.Debug {
		// Restore body so DumpRequest can include it.
		r.Body = io.NopCloser(bytes.NewReader(body))

		if dump, err := httputil.DumpRequest(r, true); err == nil {
			s.Log.Info.Println("=========== INCOMING REQUEST START ===========")
			s.Log.Info.Println(string(dump))
			s.Log.Info.Println("=========== INCOMING REQUEST END =============")
		} else {
			s.Log.Err.Printf("dump request failed: %v", err)
		}

		// Pretty-print JSON if possible (helps with debugging payload shape).
		var pretty bytes.Buffer
		if err := json.Indent(&pretty, body, "", "  "); err == nil {
			s.Log.Info.Println("=========== JSON PRETTY START ================")
			s.Log.Info.Println(pretty.String())
			s.Log.Info.Println("=========== JSON PRETTY END ==================")
		}
	}

	// Parse Alertmanager webhook JSON payload.
	var wh ocp.Webhook
	if err := json.Unmarshal(body, &wh); err != nil {
		s.Log.Err.Printf("json unmarshal failed: %v", err)
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	// Minimal receive log for production mode (Debug=false).
	s.Log.Info.Printf(
		"alertmanager webhook received status=%s receiver=%s alerts=%d truncated=%d",
		wh.Status, wh.Receiver, len(wh.Alerts), wh.TruncatedAlerts,
	)

	// Build Synology message text from Alertmanager payload.
	text := api.FormatSynologyText(wh)

	// Debug mode: log the exact text to be sent to Synology.
	if s.Debug {
		s.Log.Info.Println("=========== SYNOLOGY TEXT START ==============")
		s.Log.Info.Println(text)
		s.Log.Info.Println("=========== SYNOLOGY TEXT END ================")
	}

	// Send formatted message to Synology Chat if configured.
	if s.Sender != nil && s.Sender.WebhookURL != "" {
		if err := s.Sender.SendText(text, s.Debug, s.Log); err != nil {
			s.Log.Err.Printf("synology send failed: %v", err)
			// Typically we still return 200 OK to avoid Alertmanager retry storms.
		} else {
			s.Log.Info.Println("synology chat sent OK")
		}
	} else {
		// Only log this in debug to avoid noise in production logs.
		if s.Debug {
			s.Log.Info.Println("synology sender not configured (disabled or empty webhook_url)")
		}
	}

	// Always respond OK once the webhook has been processed.
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok\n"))
}
