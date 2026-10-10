package tgnotify

import (
	"net/http"
	"time"
)

// Config defines the configuration for Notifier.
type Config struct {
	BotToken    string
	ChatID      string
	ThreadID    *int
	BaseURL     string
	AppName     string
	Environment string
	Timeout     time.Duration
	HTTPClient  *http.Client
	Async       bool
	OnError     func(err error)
}

// Option configures the Notifier.
type Option func(*Config)

// WithBaseURL sets a custom Telegram API base URL (useful for local mock tests or proxies).
func WithBaseURL(url string) Option {
	return func(c *Config) {
		c.BaseURL = url
	}
}

// WithThreadID sets the default message thread ID for Telegram topics/supergroups.
func WithThreadID(threadID int) Option {
	return func(c *Config) {
		c.ThreadID = &threadID
	}
}

// WithAppName sets the application name included in notification headers.
func WithAppName(name string) Option {
	return func(c *Config) {
		c.AppName = name
	}
}

// WithEnvironment sets the application environment (e.g., "production", "staging").
func WithEnvironment(env string) Option {
	return func(c *Config) {
		c.Environment = env
	}
}

// WithTimeout sets the HTTP timeout for Telegram requests (default: 10s).
func WithTimeout(d time.Duration) Option {
	return func(c *Config) {
		if d > 0 {
			c.Timeout = d
		}
	}
}

// WithHTTPClient allows passing a custom *http.Client.
func WithHTTPClient(client *http.Client) Option {
	return func(c *Config) {
		if client != nil {
			c.HTTPClient = client
		}
	}
}

// WithAsync enables sending messages in background goroutines.
func WithAsync(enabled bool) Option {
	return func(c *Config) {
		c.Async = enabled
	}
}

// WithErrorHandler sets an error handler callback (useful in async mode).
func WithErrorHandler(fn func(err error)) Option {
	return func(c *Config) {
		c.OnError = fn
	}
}

// MessageOption provides per-message customizations.
type MessageOption func(*messageSettings)

type messageSettings struct {
	ChatID   string
	ThreadID *int
	Extra    map[string]any
}

// PerThreadID overrides the thread ID for a specific notification.
func PerThreadID(threadID int) MessageOption {
	return func(s *messageSettings) {
		s.ThreadID = &threadID
	}
}

// PerChatID overrides the chat ID for a specific notification.
func PerChatID(chatID string) MessageOption {
	return func(s *messageSettings) {
		s.ChatID = chatID
	}
}

// WithExtra adds metadata key-values to the notification.
func WithExtra(extra map[string]any) MessageOption {
	return func(s *messageSettings) {
		if s.Extra == nil {
			s.Extra = make(map[string]any)
		}
		for k, v := range extra {
			s.Extra[k] = v
		}
	}
}
