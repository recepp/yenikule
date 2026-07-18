package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"yenikule/config"
)

// ── Pool: reuse byte buffers so form payloads don't pressure the GC ──────────

var bufPool = sync.Pool{
	New: func() any { return new(bytes.Buffer) },
}

// ── Telegram API types (only the fields we actually use) ─────────────────────

// tgUpdate is the minimal shape of a Telegram Update object.
type tgUpdate struct {
	Message *tgMessage `json:"message"`
}

type tgMessage struct {
	Chat tgChat `json:"chat"`
	Text string `json:"text"`
}

type tgChat struct {
	ID int64 `json:"id"`
}

// tgSend is the body we POST to sendMessage.
type tgSend struct {
	ChatID    string `json:"chat_id"`
	Text      string `json:"text"`
	ParseMode string `json:"parse_mode"`
}

// ── Shared HTTP client — one instance, keeps connection pool alive ────────────

var tgClient = &http.Client{
	Timeout: 8 * time.Second,
	Transport: &http.Transport{
		MaxIdleConnsPerHost: 1, // bot only talks to one host
	},
}

// ── Handler registration ──────────────────────────────────────────────────────

// RegisterTelegram attaches the webhook endpoint to mux.
// The path includes a random secret so it is not guessable.
func RegisterTelegram(mux *http.ServeMux, cfg *config.Config) {
	apiBase := fmt.Sprintf("https://api.telegram.org/bot%s", cfg.TelegramToken)

	mux.HandleFunc("/webhook/"+cfg.WebhookPath, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Limit body to 64 KB — webhook payloads are tiny
		r.Body = http.MaxBytesReader(w, r.Body, 64<<10)

		buf := bufPool.Get().(*bytes.Buffer)
		buf.Reset()
		defer bufPool.Put(buf)

		if _, err := io.Copy(buf, r.Body); err != nil {
			http.Error(w, "read error", http.StatusBadRequest)
			return
		}

		var update tgUpdate
		if err := json.Unmarshal(buf.Bytes(), &update); err != nil || update.Message == nil {
			// Telegram retries on non-2xx; acknowledge silently for unknown shapes
			w.WriteHeader(http.StatusOK)
			return
		}

		text := strings.TrimSpace(update.Message.Text)
		if text == "" {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Echo / relay to configured chat — fire and forget to keep latency low
		go sendMessage(apiBase, cfg.TelegramChatID, text)

		w.WriteHeader(http.StatusOK)
	})

	// Contact-form endpoint: /api/contact
	// The JS form POSTs here; handler relays the message to Telegram.
	mux.HandleFunc("/api/contact", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		r.Body = http.MaxBytesReader(w, r.Body, 32<<10)
		if err := r.ParseForm(); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}

		name    := strings.TrimSpace(r.FormValue("name"))
		email   := strings.TrimSpace(r.FormValue("email"))
		subject := strings.TrimSpace(r.FormValue("subject"))
		message := strings.TrimSpace(r.FormValue("message"))

		if name == "" || email == "" || message == "" {
			http.Error(w, "name, email and message are required", http.StatusUnprocessableEntity)
			return
		}

		text := fmt.Sprintf(
			"📬 *Yeni İletişim Formu*\n\n"+
				"👤 Ad: %s\n"+
				"📧 E-posta: %s\n"+
				"📌 Konu: %s\n\n"+
				"💬 Mesaj:\n%s",
			escMD(name), escMD(email), escMD(subject), escMD(message),
		)

		go sendMessage(apiBase, cfg.TelegramChatID, text)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":true}`))
	})
}

// sendMessage posts a Telegram sendMessage call.
// Called from a goroutine — does not block the HTTP response.
func sendMessage(apiBase, chatID, text string) {
	payload := tgSend{
		ChatID:    chatID,
		Text:      text,
		ParseMode: "Markdown",
	}

	body, err := json.Marshal(payload)
	if err != nil {
		log.Printf("[telegram] marshal error: %v", err)
		return
	}

	resp, err := tgClient.Post(
		apiBase+"/sendMessage",
		"application/json",
		bytes.NewReader(body),
	)
	if err != nil {
		log.Printf("[telegram] send error: %v", err)
		return
	}
	defer resp.Body.Close()
	// Drain body to reuse the TCP connection
	_, _ = io.Copy(io.Discard, resp.Body)
}

// escMD escapes characters that have special meaning in Telegram Markdown v1.
func escMD(s string) string {
	r := strings.NewReplacer(
		"_", "\\_",
		"*", "\\*",
		"`", "\\`",
		"[", "\\[",
	)
	return r.Replace(s)
}
