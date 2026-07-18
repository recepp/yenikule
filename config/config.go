package config

import (
	"crypto/rand"
	"encoding/hex"
	"os"
)

// Config holds all runtime settings.
// All fields are populated once at startup — zero per-request allocation.
type Config struct {
	Addr           string // e.g. ":8080"
	StaticDir      string // path to the static files root
	TelegramToken  string // bot token; empty → webhook disabled
	TelegramChatID string // target chat/group ID for form submissions
	WebhookPath    string // random token appended to /webhook/ for security
}

// Load reads config from environment variables with sensible defaults.
func Load() *Config {
	cfg := &Config{
		Addr:           getEnv("PORT", ":8080"),
		StaticDir:      getEnv("STATIC_DIR", "static"),
		TelegramToken:  os.Getenv("TELEGRAM_TOKEN"),
		TelegramChatID: os.Getenv("TELEGRAM_CHAT_ID"),
	}

	// Normalise PORT: if value has no colon prefix add one
	if len(cfg.Addr) > 0 && cfg.Addr[0] != ':' {
		cfg.Addr = ":" + cfg.Addr
	}

	// Generate a random webhook path so the endpoint is not guessable.
	// This is cheaper and safer than a config value that might be leaked.
	cfg.WebhookPath = randomHex(16)

	return cfg
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func randomHex(n int) string {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		panic("config: cannot generate webhook secret: " + err.Error())
	}
	return hex.EncodeToString(b)
}
