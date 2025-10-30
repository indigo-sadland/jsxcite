package filter

import (
	"regexp"
)

// IsMimeType filters out strings that are valid http content types
func IsMimeType(content string) bool {

	patterns := compileMIMEPatterns()

	// Check against all patterns
	for _, pattern := range patterns {
		if pattern.MatchString(content) {
			return true
		}
	}

	return false
}

// IsURL checks if a string matches URL/URI patterns
func IsURL(content string) bool {

	patterns := compileURLPatterns()

	// Skip very short strings or strings with spaces
	if len(content) < 3 || regexp.MustCompile(`\s`).MatchString(content) {
		return false
	}

	// Check against MIME patterns
	if IsMimeType(content) {
		return false
	}

	// Check against all patterns
	for _, pattern := range patterns {
		if pattern.MatchString(content) {
			return true
		}
	}

	return false
}

// compileURLPatterns creates regex patterns for different URL/URI types
func compileURLPatterns() []*regexp.Regexp {
	patterns := []string{
		// Full URLs with protocol
		`^(https?|wss?|ftp|file)://`,

		// Absolute paths (starting with /)
		`^/[a-zA-Z0-9_\-./]+`,

		// Relative paths with multiple segments containing /
		`^[a-zA-Z0-9_\-]+(/[a-zA-Z0-9_\-./]+)+`,

		// Relative parent paths
		`^\.\.?/`,

		`(?:localhost|\b\d{1,3}(?:\.\d{1,3}){3}\b)(?::\d{1,5})?`,
	}

	var compiled []*regexp.Regexp
	for _, pattern := range patterns {
		compiled = append(compiled, regexp.MustCompile(pattern))
	}
	return compiled
}

// compileMIMEPatterns creates regex patterns for different HTTP MIME types
func compileMIMEPatterns() []*regexp.Regexp {
	patterns := []string{
		`^(application|text|image|audio|video|font|multipart|message|model)/[a-zA-Z0-9][a-zA-Z0-9\-+.]*`,
	}

	var compiled []*regexp.Regexp
	for _, pattern := range patterns {
		compiled = append(compiled, regexp.MustCompile(pattern))
	}
	return compiled
}
