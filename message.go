package tgnotify

import (
	"fmt"
	"html"
	"runtime"
	"sort"
	"strings"
	"time"
)

const maxTelegramMessageLength = 4000

// EscapeHTML escapes special HTML characters for Telegram HTML mode.
func EscapeHTML(s string) string {
	return html.EscapeString(s)
}

// TruncateMessage ensures message length does not exceed Telegram limits.
func TruncateMessage(msg string) string {
	if len(msg) <= maxTelegramMessageLength {
		return msg
	}
	return msg[:maxTelegramMessageLength-30] + "\n<i>...(pesan dipotong)...</i>"
}

// FormatErrorMessage builds an HTML message for system errors / exceptions.
func FormatErrorMessage(cfg Config, err error, callerFile string, callerLine int, callerFunc string, extra map[string]any) string {
	var b strings.Builder

	header := "🚨 <b>SYSTEM ERROR / EXCEPTION</b>"
	if cfg.AppName != "" || cfg.Environment != "" {
		tags := ""
		if cfg.AppName != "" {
			tags += "[" + EscapeHTML(cfg.AppName) + "]"
		}
		if cfg.Environment != "" {
			tags += " [" + strings.ToUpper(EscapeHTML(cfg.Environment)) + "]"
		}
		header = fmt.Sprintf("🚨 <b>%s SYSTEM ERROR</b>", tags)
	}

	b.WriteString(header + "\n\n")

	if err != nil {
		b.WriteString(fmt.Sprintf("<b>Error:</b>\n<code>%s</code>\n\n", EscapeHTML(err.Error())))
	}

	if callerFile != "" || callerFunc != "" {
		b.WriteString("<b>Location:</b>\n")
		if callerFunc != "" {
			b.WriteString(fmt.Sprintf("• Function: <code>%s</code>\n", EscapeHTML(callerFunc)))
		}
		if callerFile != "" {
			b.WriteString(fmt.Sprintf("• File: <code>%s:%d</code>\n", EscapeHTML(callerFile), callerLine))
		}
		b.WriteString("\n")
	}

	if len(extra) > 0 {
		b.WriteString("<b>Metadata:</b>\n")
		keys := make([]string, 0, len(extra))
		for k := range extra {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		for _, k := range keys {
			b.WriteString(fmt.Sprintf("• <b>%s:</b> %s\n", EscapeHTML(k), EscapeHTML(fmt.Sprintf("%v", extra[k]))))
		}
		b.WriteString("\n")
	}

	b.WriteString(fmt.Sprintf("🕒 <i>%s</i>", time.Now().Format("2006-01-02 15:04:05 MST")))

	return TruncateMessage(b.String())
}

// FormatAlertMessage builds an HTML message for alerts, third-party failures, or operational notices.
func FormatAlertMessage(cfg Config, title string, detail string, extra map[string]any) string {
	var b strings.Builder

	header := "⚠️ <b>ALERT NOTIFICATION</b>"
	if cfg.AppName != "" || cfg.Environment != "" {
		tags := ""
		if cfg.AppName != "" {
			tags += "[" + EscapeHTML(cfg.AppName) + "]"
		}
		if cfg.Environment != "" {
			tags += " [" + strings.ToUpper(EscapeHTML(cfg.Environment)) + "]"
		}
		header = fmt.Sprintf("⚠️ <b>%s %s</b>", tags, EscapeHTML(title))
	} else if title != "" {
		header = fmt.Sprintf("⚠️ <b>%s</b>", EscapeHTML(title))
	}

	b.WriteString(header + "\n\n")

	if detail != "" {
		b.WriteString(fmt.Sprintf("<b>Detail:</b>\n%s\n\n", EscapeHTML(detail)))
	}

	if len(extra) > 0 {
		b.WriteString("<b>Metadata:</b>\n")
		keys := make([]string, 0, len(extra))
		for k := range extra {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		for _, k := range keys {
			b.WriteString(fmt.Sprintf("• <b>%s:</b> %s\n", EscapeHTML(k), EscapeHTML(fmt.Sprintf("%v", extra[k]))))
		}
		b.WriteString("\n")
	}

	b.WriteString(fmt.Sprintf("🕒 <i>%s</i>", time.Now().Format("2006-01-02 15:04:05 MST")))

	return TruncateMessage(b.String())
}

// captureCallerFrame extracts the caller frame skipping library internal frames.
func captureCallerFrame(skip int) (file string, line int, fnName string) {
	pc, file, line, ok := runtime.Caller(skip)
	if !ok {
		return "", 0, ""
	}
	fn := runtime.FuncForPC(pc)
	if fn != nil {
		fullName := fn.Name()
		lastSlash := strings.LastIndex(fullName, "/")
		if lastSlash >= 0 {
			fullName = fullName[lastSlash+1:]
		}
		fnName = fullName
	}
	return file, line, fnName
}
