package validators

import (
	"fmt"
	"html"
	"regexp"
	"strings"
	"unicode"
)

// SanitizeHTML removes potentially dangerous HTML/JavaScript
func SanitizeHTML(input string) string {
	// Escape HTML entities
	sanitized := html.EscapeString(input)

	// Remove script tags and content
	scriptRegex := regexp.MustCompile(`(?i)<script[^>]*>.*?</script>`)
	sanitized = scriptRegex.ReplaceAllString(sanitized, "")

	// Remove event handlers
	eventRegex := regexp.MustCompile(`(?i)on\w+\s*=\s*["'][^"']*["']`)
	sanitized = eventRegex.ReplaceAllString(sanitized, "")

	// Remove javascript: protocol
	jsRegex := regexp.MustCompile(`(?i)javascript:`)
	sanitized = jsRegex.ReplaceAllString(sanitized, "")

	return sanitized
}

// SanitizeSQL prevents SQL injection by escaping special characters
// Note: Use parameterized queries as primary defense
func SanitizeSQL(input string) string {
	// Escape single quotes
	sanitized := strings.ReplaceAll(input, "'", "''")

	// Remove SQL comments
	sanitized = regexp.MustCompile(`--.*$`).ReplaceAllString(sanitized, "")
	sanitized = regexp.MustCompile(`/\*.*?\*/`).ReplaceAllString(sanitized, "")

	// Remove dangerous SQL keywords (defense in depth)
	dangerousPatterns := []string{
		`;\s*DROP`,
		`;\s*DELETE`,
		`;\s*UPDATE`,
		`;\s*INSERT`,
		`;\s*EXEC`,
		`;\s*EXECUTE`,
		`UNION\s+SELECT`,
	}

	for _, pattern := range dangerousPatterns {
		regex := regexp.MustCompile(`(?i)` + pattern)
		if regex.MatchString(sanitized) {
			// If dangerous pattern detected, escape the entire string
			return regexp.QuoteMeta(sanitized)
		}
	}

	return sanitized
}

// SanitizeFileName removes dangerous characters from filenames
func SanitizeFileName(filename string) string {
	// Remove path traversal attempts
	filename = strings.ReplaceAll(filename, "..", "")
	filename = strings.ReplaceAll(filename, "/", "")
	filename = strings.ReplaceAll(filename, "\\", "")

	// Remove null bytes
	filename = strings.ReplaceAll(filename, "\x00", "")

	// Allow only alphanumeric, underscore, dash, and dot
	validChars := regexp.MustCompile(`[^a-zA-Z0-9_\-\.]`)
	filename = validChars.ReplaceAllString(filename, "_")

	// Limit length
	if len(filename) > 255 {
		filename = filename[:255]
	}

	return filename
}

// TrimAndNormalize trims whitespace and normalizes spacing
func TrimAndNormalize(input string) string {
	// Trim leading/trailing whitespace
	trimmed := strings.TrimSpace(input)

	// Replace multiple spaces with single space
	spaceRegex := regexp.MustCompile(`\s+`)
	trimmed = spaceRegex.ReplaceAllString(trimmed, " ")

	return trimmed
}

// RemoveNonPrintable removes non-printable characters
func RemoveNonPrintable(input string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsPrint(r) {
			return r
		}
		return -1
	}, input)
}

// SanitizeURL validates and sanitizes URL
func SanitizeURL(url string) string {
	url = strings.TrimSpace(url)

	// Remove javascript: and data: protocols
	dangerousProtocols := []string{
		"javascript:",
		"data:",
		"vbscript:",
		"file:",
	}

	urlLower := strings.ToLower(url)
	for _, protocol := range dangerousProtocols {
		if strings.HasPrefix(urlLower, protocol) {
			return ""
		}
	}

	// Only allow http and https
	if !strings.HasPrefix(urlLower, "http://") && !strings.HasPrefix(urlLower, "https://") {
		// Add https if no protocol specified
		if !strings.Contains(url, "://") {
			url = "https://" + url
		} else {
			return "" // Invalid protocol
		}
	}

	return url
}

// SanitizeJSONString sanitizes string for JSON output
func SanitizeJSONString(input string) string {
	// Escape special JSON characters
	input = strings.ReplaceAll(input, "\\", "\\\\")
	input = strings.ReplaceAll(input, "\"", "\\\"")
	input = strings.ReplaceAll(input, "\n", "\\n")
	input = strings.ReplaceAll(input, "\r", "\\r")
	input = strings.ReplaceAll(input, "\t", "\\t")

	return input
}

// CleanPhoneNumber removes formatting from phone number
func CleanPhoneNumber(phone string) string {
	// Remove common formatting characters
	phone = strings.ReplaceAll(phone, " ", "")
	phone = strings.ReplaceAll(phone, "-", "")
	phone = strings.ReplaceAll(phone, "(", "")
	phone = strings.ReplaceAll(phone, ")", "")
	phone = strings.ReplaceAll(phone, ".", "")

	return strings.TrimSpace(phone)
}

// StripHTML removes all HTML tags
func StripHTML(input string) string {
	// Remove all HTML tags
	htmlRegex := regexp.MustCompile(`<[^>]*>`)
	stripped := htmlRegex.ReplaceAllString(input, "")

	// Unescape HTML entities
	stripped = html.UnescapeString(stripped)

	return strings.TrimSpace(stripped)
}

// LimitLength limits string length and adds ellipsis if truncated
func LimitLength(input string, maxLength int) string {
	if len(input) <= maxLength {
		return input
	}

	if maxLength < 3 {
		return input[:maxLength]
	}

	return input[:maxLength-3] + "..."
}

// SanitizeSearchQuery sanitizes user search input
func SanitizeSearchQuery(query string) string {
	// Trim and normalize
	query = TrimAndNormalize(query)

	// Remove SQL wildcards that could cause DoS
	query = strings.ReplaceAll(query, "%", "")
	query = strings.ReplaceAll(query, "_", "")

	// Remove special regex characters
	specialChars := []string{"*", "+", "?", "[", "]", "{", "}", "(", ")", "|", "^", "$", "\\"}
	for _, char := range specialChars {
		query = strings.ReplaceAll(query, char, "")
	}

	// Limit length
	if len(query) > 100 {
		query = query[:100]
	}

	return query
}

// SanitizeNumericString ensures string contains only digits
func SanitizeNumericString(input string) string {
	return regexp.MustCompile(`[^0-9]`).ReplaceAllString(input, "")
}

// SanitizeAlphanumeric ensures string contains only letters and numbers
func SanitizeAlphanumeric(input string) string {
	return regexp.MustCompile(`[^a-zA-Z0-9]`).ReplaceAllString(input, "")
}

// RemoveInvisibleChars removes zero-width and other invisible characters
func RemoveInvisibleChars(input string) string {
	invisibleChars := []rune{
		'\u200B', // Zero width space
		'\u200C', // Zero width non-joiner
		'\u200D', // Zero width joiner
		'\uFEFF', // Zero width no-break space
		'\u00AD', // Soft hyphen
	}

	for _, char := range invisibleChars {
		input = strings.ReplaceAll(input, string(char), "")
	}

	return input
}

// NormalizeWhitespace normalizes whitespace characters
func NormalizeWhitespace(input string) string {
	// Replace tabs, newlines, etc. with spaces
	input = strings.ReplaceAll(input, "\t", " ")
	input = strings.ReplaceAll(input, "\n", " ")
	input = strings.ReplaceAll(input, "\r", " ")

	// Replace multiple spaces with single space
	spaceRegex := regexp.MustCompile(`\s+`)
	input = spaceRegex.ReplaceAllString(input, " ")

	return strings.TrimSpace(input)
}

// SanitizeCoordinates validates and sanitizes GPS coordinates
func SanitizeCoordinates(lat, lng float64) (float64, float64, error) {
	// Validate latitude (-90 to 90)
	if lat < -90 || lat > 90 {
		return 0, 0, fmt.Errorf("latitude must be between -90 and 90")
	}

	// Validate longitude (-180 to 180)
	if lng < -180 || lng > 180 {
		return 0, 0, fmt.Errorf("longitude must be between -180 and 180")
	}

	// Check if coordinates are in Indonesia (rough bounds)
	// Indonesia: approximately 6°N to 11°S, 95°E to 141°E
	if lat < -11 || lat > 6 || lng < 95 || lng > 141 {
		// Still valid coordinates, but warn about non-Indonesian location
		// You might want to add a warning log here
	}

	return lat, lng, nil
}

// Sanitize struct provides a collection of sanitization methods
type Sanitizer struct{}

// NewSanitizer creates a new sanitizer
func NewSanitizer() *Sanitizer {
	return &Sanitizer{}
}

// SanitizeInput applies multiple sanitization rules
func (s *Sanitizer) SanitizeInput(input string) string {
	// Remove invisible characters
	input = RemoveInvisibleChars(input)

	// Remove non-printable characters
	input = RemoveNonPrintable(input)

	// Normalize whitespace
	input = NormalizeWhitespace(input)

	// Trim
	input = strings.TrimSpace(input)

	return input
}

// SanitizeUserInput sanitizes user-provided text input
func (s *Sanitizer) SanitizeUserInput(input string, maxLength int) string {
	// Apply basic sanitization
	input = s.SanitizeInput(input)

	// Remove HTML
	input = StripHTML(input)

	// Limit length
	if maxLength > 0 {
		input = LimitLength(input, maxLength)
	}

	return input
}

