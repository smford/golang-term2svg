package colorize

import (
	"regexp"
	"strings"
)

// ANSI color sequences matching tui-logs.svg aesthetic
const (
	ansiReset      = "\x1b[0m"
	ansiDim        = "\x1b[2;37m"       // dark dim timestamp #64748b
	ansiBoldGreen  = "\x1b[1;32m"       // [INFO], [OK], 200, Running, healthy #10b981
	ansiBoldYellow = "\x1b[1;33m"       // [WARN], Degraded, Pending #f59e0b
	ansiBoldRed    = "\x1b[1;31m"       // [ERROR], [FATAL], Fail, CrashLoopBackOff #ef4444
	ansiBoldBlue   = "\x1b[1;34m"       // Prompts, commands
	ansiCyan       = "\x1b[36m"         // key= in key=value pairs #38bdf8
	ansiBoldPurple = "\x1b[1;35m"       // Headers, decorative frames
)

type patternRule struct {
	re    *regexp.Regexp
	color string
}

var (
	// Timestamps: 14:03:00.012 or 2026-10-10 22:30:00 or 2026-10-10T22:30:00Z
	reTimestamp = regexp.MustCompile(`\b\d{2}:\d{2}:\d{2}(?:\.\d{3,6})?\b|\b\d{4}-\d{2}-\d{2}[T ]\d{2}:\d{2}:\d{2}(?:\.\d{3,6})?(?:Z|[+-]\d{2}:\d{2})?\b`)

	// Status levels
	reInfo = regexp.MustCompile(`(?i)\b(?:\[INFO\s*\]|INFO|Running|Completed|Healthy|Success|SUCCESS|OK|200 OK|200 GET|200 POST|200)\b|✔`)
	reWarn = regexp.MustCompile(`(?i)\b(?:\[WARN\s*\]|\[WARNING\s*\]|WARN|WARNING|Degraded|Caution|Pending)\b|▲`)
	reErr  = regexp.MustCompile(`(?i)\b(?:\[ERROR\s*\]|\[FATAL\s*\]|ERROR|FATAL|CrashLoopBackOff|Failed|Failure|CRITICAL|Panic|500 Internal Server Error|500)\b|✖`)

	// Key-value pairs: version=v0.2.0, fleet=4, client=192.168.1.1:80
	reKeyValue = regexp.MustCompile(`\b([a-zA-Z0-9_-]+)=([^\s]+)`)

	// Prompt headers: ➜ ~ or $ command or > prompt
	rePrompt = regexp.MustCompile(`^(\s*(?:➜|❯|\$|>|●|○)\s*)`)
)

// Line transforms a plain-text terminal line into an ANSI colorized line matching tui-logs.svg.
func Line(line string) string {
	// If already contains ANSI escapes, return untouched
	if strings.Contains(line, "\x1b[") {
		return line
	}

	trimmed := strings.TrimRight(line, "\r\n")
	if len(strings.TrimSpace(trimmed)) == 0 {
		return line
	}

	result := trimmed

	// 1. Colorize Key-Value pairs: key=value -> cyan(key)=value
	result = reKeyValue.ReplaceAllString(result, ansiCyan+"$1"+ansiReset+"=$2")

	// 2. Colorize Timestamps
	result = reTimestamp.ReplaceAllStringFunc(result, func(m string) string {
		return ansiDim + m + ansiReset
	})

	// 3. Colorize Status Tokens
	result = reInfo.ReplaceAllStringFunc(result, func(m string) string {
		return ansiBoldGreen + m + ansiReset
	})
	result = reWarn.ReplaceAllStringFunc(result, func(m string) string {
		return ansiBoldYellow + m + ansiReset
	})
	result = reErr.ReplaceAllStringFunc(result, func(m string) string {
		return ansiBoldRed + m + ansiReset
	})

	// 4. Colorize Prompt bullet or prefix
	if rePrompt.MatchString(result) {
		result = rePrompt.ReplaceAllString(result, ansiBoldBlue+"$1"+ansiReset)
	}

	return result + "\n"
}

// Text colorizes multi-line plain text strings.
func Text(text string) string {
	lines := strings.Split(text, "\n")
	var sb strings.Builder
	for i, l := range lines {
		if i == len(lines)-1 && len(l) == 0 {
			break
		}
		sb.WriteString(Line(l))
	}
	return sb.String()
}
