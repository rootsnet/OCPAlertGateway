package main

import (
	"flag"
	"net/http"
	"net/url"
	"strings"
	"time"

	"ocp-alert-gateway/config"
	"ocp-alert-gateway/logger"
	"ocp-alert-gateway/synologychat"
	"ocp-alert-gateway/webserver"
)

// maskWebhook masks sensitive token information from a Synology webhook URL.
// If a "token" query parameter exists, only the first 4 characters are kept.
// If no token exists and the URL is too long, it is truncated for safety.
func maskWebhook(u string) string {
	pu, err := url.Parse(strings.TrimSpace(u))
	if err != nil {
		return "(invalid url)"
	}

	// Mask token if present in query parameters
	q := pu.Query()
	if tok := q.Get("token"); tok != "" {
		keep := tok
		if len(tok) > 4 {
			keep = tok[:4]
		}
		q.Set("token", keep+"***")
		pu.RawQuery = q.Encode()
		return pu.String()
	}

	// If no token parameter exists, shorten very long URLs
	s := pu.String()
	if len(s) > 120 {
		return s[:120] + "..."
	}
	return s
}

func main() {
	// Parse command-line flags
	var configPath string
	flag.StringVar(&configPath, "config", "config/config.yaml", "config file path")
	flag.Parse()

	// Initialize logger
	log := logger.New()

	// Load configuration file
	cfg, err := config.Load(configPath)
	if err != nil {
		log.Err.Fatalf("load config failed: %v", err)
	}

	// Prepare Synology Chat sender if enabled and webhook URL is provided
	var sender *synologychat.Sender
	if cfg.SynologyChat.Enabled && strings.TrimSpace(cfg.SynologyChat.WebhookURL) != "" {
		if cfg.Debug {
			log.Info.Printf("synology webhook url=%s", maskWebhook(cfg.SynologyChat.WebhookURL))
		}
		sender = &synologychat.Sender{
			WebhookURL:         cfg.SynologyChat.WebhookURL,
			InsecureSkipVerify: cfg.SynologyChat.InsecureSkipVerify,
		}
	} else {
		if cfg.Debug {
			log.Info.Printf("synology disabled or webhook_url empty (enabled=%v)", cfg.SynologyChat.Enabled)
		}
	}

	// Create web server instance
	srv := &webserver.Server{
		Log:    log,
		Sender: sender,
		Debug:  cfg.Debug,
	}

	// HTTP routing setup
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", srv.Healthz)
	mux.HandleFunc(cfg.Server.WebhookPath, srv.Webhook)

	// HTTP server configuration
	server := &http.Server{
		Addr:              cfg.Server.ListenAddr,
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
	}

	// Startup log
	log.Info.Printf("listening on %s (path=%s) debug=%v config=%s",
		cfg.Server.ListenAddr, cfg.Server.WebhookPath, cfg.Debug, configPath)

	// Start HTTP server (blocks until error)
	log.Err.Fatal(server.ListenAndServe())
}
