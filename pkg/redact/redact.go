package redact

import (
	"regexp"
)

// Rule represents a redaction rule with a regex pattern and replacement string.
type Rule struct {
	Name        string
	Pattern     *regexp.Regexp
	Replacement string
}

// DefaultRules returns industry-standard SRE & security redaction rules.
func DefaultRules() []Rule {
	return []Rule{
		{
			Name:        "IPv4",
			Pattern:     regexp.MustCompile(`\b(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\b`),
			Replacement: "[REDACTED_IP]",
		},
		{
			Name:        "BearerToken",
			Pattern:     regexp.MustCompile(`(?i)(bearer\s+)[a-zA-Z0-9_\-\.]{15,}`),
			Replacement: `$1[REDACTED_TOKEN]`,
		},
		{
			Name:        "APIKey",
			Pattern:     regexp.MustCompile(`(?i)(api[_-]?key[:=]\s*)[a-zA-Z0-9_\-]{16,}`),
			Replacement: `$1[REDACTED_KEY]`,
		},
		{
			Name:        "JWT",
			Pattern:     regexp.MustCompile(`\beyJ[a-zA-Z0-9_-]+\.eyJ[a-zA-Z0-9_-]+\.[a-zA-Z0-9_-]+\b`),
			Replacement: "[REDACTED_JWT]",
		},
		{
			Name:        "Password",
			Pattern:     regexp.MustCompile(`(?i)(password[:=]\s*)[^\s]+`),
			Replacement: `$1[REDACTED_PASSWORD]`,
		},
	}
}

// Redactor applies sanitization rules to text.
type Redactor struct {
	rules []Rule
}

// New creates a new Redactor with optional custom rules.
func New(rules ...Rule) *Redactor {
	if len(rules) == 0 {
		rules = DefaultRules()
	}
	return &Redactor{rules: rules}
}

// Apply transforms input text by masking all matched patterns.
func (r *Redactor) Apply(text string) string {
	res := text
	for _, rule := range r.rules {
		res = rule.Pattern.ReplaceAllString(res, rule.Replacement)
	}
	return res
}
