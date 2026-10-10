package tgnotify

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const defaultTelegramBaseURL = "https://api.telegram.org"

// Notifier sends messages and alerts to a Telegram chat or thread.
type Notifier struct {
	cfg        Config
	httpClient *http.Client
}

// New initializes a new Telegram Notifier.
// botToken and chatID are required.
func New(botToken, chatID string, opts ...Option) (*Notifier, error) {
	botToken = strings.TrimSpace(botToken)
	chatID = strings.TrimSpace(chatID)

	if botToken == "" {
		return nil, errors.New("tgnotify: botToken cannot be empty")
	}
	if chatID == "" {
		return nil, errors.New("tgnotify: chatID cannot be empty")
	}

	cfg := Config{
		BotToken: botToken,
		ChatID:   chatID,
		BaseURL:  defaultTelegramBaseURL,
		Timeout:  10 * time.Second,
		OnError:  func(err error) {},
	}

	for _, opt := range opts {
		opt(&cfg)
	}

	httpClient := cfg.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{
			Timeout: cfg.Timeout,
		}
	}

	return &Notifier{
		cfg:        cfg,
		httpClient: httpClient,
	}, nil
}

type telegramPayload struct {
	ChatID                string `json:"chat_id"`
	MessageThreadID       *int   `json:"message_thread_id,omitempty"`
	Text                  string `json:"text"`
	ParseMode             string `json:"parse_mode,omitempty"`
	DisableWebPagePreview bool   `json:"disable_web_page_preview"`
}

type telegramResponse struct {
	OK          bool   `json:"ok"`
	Description string `json:"description"`
}

func (n *Notifier) applyMessageSettings(opts ...MessageOption) messageSettings {
	settings := messageSettings{
		ChatID:   n.cfg.ChatID,
		ThreadID: n.cfg.ThreadID,
		Extra:    make(map[string]any),
	}
	for _, opt := range opts {
		opt(&settings)
	}
	return settings
}

func (n *Notifier) sendPayload(ctx context.Context, payload telegramPayload) error {
	if n.cfg.Async {
		go func(p telegramPayload) {
			if err := n.doRequest(context.Background(), p); err != nil {
				n.cfg.OnError(fmt.Errorf("tgnotify: async send failed: %w", err))
			}
		}(payload)
		return nil
	}

	return n.doRequest(ctx, payload)
}

func (n *Notifier) doRequest(ctx context.Context, payload telegramPayload) error {
	if ctx == nil {
		ctx = context.Background()
	}

	url := fmt.Sprintf("%s/bot%s/sendMessage", strings.TrimRight(n.cfg.BaseURL, "/"), n.cfg.BotToken)

	jsonBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("tgnotify: failed to marshal payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(jsonBytes))
	if err != nil {
		return fmt.Errorf("tgnotify: failed to build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := n.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("tgnotify: http post error: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)

	var tgResp telegramResponse
	if err := json.Unmarshal(bodyBytes, &tgResp); err == nil && !tgResp.OK {
		return fmt.Errorf("tgnotify: telegram api rejected request: %s (status %d)", tgResp.Description, resp.StatusCode)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("tgnotify: unexpected http status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	return nil
}

// NotifyError formats and sends a critical error or exception alert to Telegram.
// Automatically captures the caller frame (file, line, function) where NotifyError was invoked.
func (n *Notifier) NotifyError(ctx context.Context, err error, opts ...MessageOption) error {
	if err == nil {
		return nil
	}

	settings := n.applyMessageSettings(opts...)
	file, line, fn := captureCallerFrame(2)

	formatted := FormatErrorMessage(n.cfg, err, file, line, fn, settings.Extra)

	payload := telegramPayload{
		ChatID:                settings.ChatID,
		MessageThreadID:       settings.ThreadID,
		Text:                  formatted,
		ParseMode:             "HTML",
		DisableWebPagePreview: true,
	}

	return n.sendPayload(ctx, payload)
}

// NotifyAlert formats and sends a non-panic/non-exception alert (e.g. 3rd-party API failure, transaction issue).
func (n *Notifier) NotifyAlert(ctx context.Context, title string, detail string, opts ...MessageOption) error {
	settings := n.applyMessageSettings(opts...)
	formatted := FormatAlertMessage(n.cfg, title, detail, settings.Extra)

	payload := telegramPayload{
		ChatID:                settings.ChatID,
		MessageThreadID:       settings.ThreadID,
		Text:                  formatted,
		ParseMode:             "HTML",
		DisableWebPagePreview: true,
	}

	return n.sendPayload(ctx, payload)
}

// SendMessage sends a plain text message without markup.
func (n *Notifier) SendMessage(ctx context.Context, text string, opts ...MessageOption) error {
	settings := n.applyMessageSettings(opts...)
	payload := telegramPayload{
		ChatID:                settings.ChatID,
		MessageThreadID:       settings.ThreadID,
		Text:                  TruncateMessage(text),
		DisableWebPagePreview: true,
	}
	return n.sendPayload(ctx, payload)
}

// SendHTML sends an HTML formatted message to Telegram.
func (n *Notifier) SendHTML(ctx context.Context, htmlText string, opts ...MessageOption) error {
	settings := n.applyMessageSettings(opts...)
	payload := telegramPayload{
		ChatID:                settings.ChatID,
		MessageThreadID:       settings.ThreadID,
		Text:                  TruncateMessage(htmlText),
		ParseMode:             "HTML",
		DisableWebPagePreview: true,
	}
	return n.sendPayload(ctx, payload)
}

// SendMarkdown sends a Markdown formatted message to Telegram.
func (n *Notifier) SendMarkdown(ctx context.Context, markdownText string, opts ...MessageOption) error {
	settings := n.applyMessageSettings(opts...)
	payload := telegramPayload{
		ChatID:                settings.ChatID,
		MessageThreadID:       settings.ThreadID,
		Text:                  TruncateMessage(markdownText),
		ParseMode:             "Markdown",
		DisableWebPagePreview: true,
	}
	return n.sendPayload(ctx, payload)
}
