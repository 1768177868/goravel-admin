package utils

import (
	"regexp"
	"strings"

	"github.com/microcosm-cc/bluemonday"
)

var (
	richContentHTMLPolicy = newRichContentHTMLPolicy()

	htmlLikeContentPattern = regexp.MustCompile(`(?i)<(p|div|span|img|br|hr|h[1-6]|ul|ol|li|table|tr|td|th|thead|tbody|blockquote|pre|code|strong|em|a|iframe)[\s>/]`)

	markdownDangerousLinkPattern = regexp.MustCompile(`(?i)\]\(\s*(?:javascript|data|vbscript):[^)]*\)`)
	rawScriptTagPattern          = regexp.MustCompile(`(?is)<script\b[^>]*>.*?</script>`)
	rawIframeTagPattern          = regexp.MustCompile(`(?is)<iframe\b[^>]*>.*?</iframe>`)
)

func newRichContentHTMLPolicy() *bluemonday.Policy {
	policy := bluemonday.UGCPolicy()
	policy.AllowRelativeURLs(true)
	policy.AllowURLSchemes("http", "https", "mailto", "tel")
	policy.RequireNoFollowOnLinks(true)
	return policy
}

// IsRichContentHTML reports whether content looks like authored HTML (WangEditor output).
func IsRichContentHTML(content string) bool {
	return htmlLikeContentPattern.MatchString(content)
}

// SanitizeRichTextContent sanitizes HTML or Markdown before persistence.
func SanitizeRichTextContent(content string) string {
	content = strings.TrimSpace(content)
	if content == "" {
		return ""
	}

	if IsRichContentHTML(content) {
		return richContentHTMLPolicy.Sanitize(content)
	}

	return sanitizeMarkdownContent(content)
}

func sanitizeMarkdownContent(content string) string {
	sanitized := rawScriptTagPattern.ReplaceAllString(content, "")
	sanitized = rawIframeTagPattern.ReplaceAllString(sanitized, "")
	sanitized = markdownDangerousLinkPattern.ReplaceAllString(sanitized, "](#)")

	if strings.Contains(sanitized, "<") {
		sanitized = richContentHTMLPolicy.Sanitize(sanitized)
	}

	return sanitized
}
