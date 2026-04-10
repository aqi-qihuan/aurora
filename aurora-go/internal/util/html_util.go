package util

import (
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/microcosm-cc/bluemonday"
)

// HTMLSanitizer XSS 防护和 HTML 过滤工具
// 对标 Java 版 HtmlUtils + 敏感词过滤
var (
	htmlPolicy = bluemonday.UGCPolicy() // 允许安全的HTML标签(a/b/i/em/strong/code/pre等)
	imgPolicy  = bluemonday.NewPolicy().AllowAttrs("src", "alt", "title").OnElements("img")
)

// SanitizeHTML 过滤危险HTML标签，防止XSS攻击
// 对标 Java 版 @Xss 注解 + Jsoup.clean()
func SanitizeHTML(html string) string {
	if html == "" {
		return ""
	}
	return htmlPolicy.Sanitize(html)
}

// SanitizeRichText 富文本内容清理（保留图片、链接等）
func SanitizeRichText(html string) string {
	if html == "" {
		return ""
	}
	p := bluemonday.NewPolicy()
	p.AllowElements("br", "p", "div", "span", "h1", "h2", "h3", "h4", "h5", "h6",
		"blockquote", "pre", "code", "ul", "ol", "li", "table", "thead", "tbody",
		"tr", "th", "td", "hr")
	p.AllowAttrs("href").OnElements("a")
	p.AllowAttrs("src", "alt", "width", "height").OnElements("img")
	p.AllowAttrs("class").Globally()
	p.AllowAttrs("style").Globally()
	return p.Sanitize(html)
}

// StripTags 移除所有 HTML 标签（纯文本提取）
// 对标 Java 版 StringUtils.stripHtmlTags()
func StripTags(html string) string {
	if html == "" {
		return ""
	}
	re := regexp.MustCompile(`<[^>]*>`)
	return re.ReplaceAllString(html, "")
}

// ExtractPlainText 提取纯文本并截断（用于摘要生成）
func ExtractPlainText(html string, maxLen int) string {
	text := StripTags(html)
	text = strings.TrimSpace(text)
	if utf8.RuneCountInString(text) <= maxLen {
		return text
	}
	runes := []rune(text)
	if len(runes) > maxLen {
		runes = runes[:maxLen]
	}
	return string(runes) + "..."
}

// EscapeHTML 转义 HTML 特殊字符（防止XSS）
func EscapeHTML(s string) string {
	var sb strings.Builder
	sb.Grow(len(s))
	for _, ch := range s {
		switch ch {
		case '&':
			sb.WriteString("&amp;")
		case '<':
			sb.WriteString("&lt;")
		case '>':
			sb.WriteString("&gt;")
		case '"':
			sb.WriteString("&quot;")
		case '\'':
			sb.WriteString("&#39;")
		default:
			sb.WriteRune(ch)
		}
	}
	return sb.String()
}

// TruncateString 截断字符串到指定长度（按字符数，不切半个中文）
func TruncateString(s string, maxLen int) string {
	if utf8.RuneCountInString(s) <= maxLen {
		return s
	}
	runes := []rune(s)
	if len(runes) > maxLen {
		runes = runes[:maxLen]
	}
	return string(runes) + "..."
}

// ContainsSensitiveWord 检查文本是否包含敏感词
func ContainsSensitiveWord(text, word string) bool {
	return strings.Contains(strings.ToLower(text), strings.ToLower(word))
}

// MaskEmail 邮箱脱敏显示
// 示例: user@example.com → u***@example.com
func MaskEmail(email string) string {
	atIndex := strings.Index(email, "@")
	if atIndex <= 1 {
		return email
	}
	return email[:1] + "***" + email[atIndex:]
}
