package util

import (
	"strings"
	"testing"
)

// ===== StripTags 测试 =====

func TestStripTags(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"empty", "", ""},
		{"no tags", "hello world", "hello world"},
		{"simple tag", "<b>bold</b>", "bold"},
		{"nested tags", "<div><p>hello</p></div>", "hello"},
		{"with attributes", `<a href="https://example.com">link</a>`, "link"},
		{"self-closing", "before<br/>after", "beforeafter"},
		{"multiple tags", "<h1>Title</h1><p>Content</p>", "TitleContent"},
		{"script tag", `<script>alert('xss')</script>hello`, "alert('xss')hello"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := StripTags(tt.input)
			if got != tt.want {
				t.Errorf("StripTags(%q) = %q; want %q", tt.input, got, tt.want)
			}
		})
	}
}

// ===== EscapeHTML 测试 =====

func TestEscapeHTML(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"no special chars", "hello world", "hello world"},
		{"ampersand", "a&b", "a&amp;b"},
		{"less than", "a<b", "a&lt;b"},
		{"greater than", "a>b", "a&gt;b"},
		{"double quote", `a"b`, "a&quot;b"},
		{"single quote", "a'b", "a&#39;b"},
		{"mixed", `<script>"alert('xss')"</script>`, "&lt;script&gt;&quot;alert(&#39;xss&#39;)&quot;&lt;/script&gt;"},
		{"empty", "", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := EscapeHTML(tt.input)
			if got != tt.want {
				t.Errorf("EscapeHTML(%q) = %q; want %q", tt.input, got, tt.want)
			}
		})
	}
}

// ===== TruncateString 测试 =====

func TestTruncateString(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		maxLen int
		want   string
	}{
		{"short enough", "hello", 10, "hello"},
		{"exact length", "hello", 5, "hello"},
		{"needs truncation", "hello world", 5, "hello..."},
		{"empty", "", 5, ""},
		{"chinese", "你好世界再见", 3, "你好世..."},
		{"zero maxLen", "hello", 0, "..."},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := TruncateString(tt.input, tt.maxLen)
			if got != tt.want {
				t.Errorf("TruncateString(%q, %d) = %q; want %q", tt.input, tt.maxLen, got, tt.want)
			}
		})
	}
}

// ===== ContainsSensitiveWord 测试 =====

func TestContainsSensitiveWord(t *testing.T) {
	tests := []struct {
		name  string
		text  string
		word  string
		want  bool
	}{
		{"contains", "This is a bad word", "bad", true},
		{"not contains", "This is a good word", "bad", false},
		{"case insensitive", "BAD WORD", "bad", true},
		{"mixed case", "BaD WoRd", "bad", true},
		{"empty text", "", "bad", false},
		{"empty word", "some text", "", true}, // strings.Contains("text", "") = true
		{"substring", "badminton", "bad", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ContainsSensitiveWord(tt.text, tt.word); got != tt.want {
				t.Errorf("ContainsSensitiveWord(%q, %q) = %v; want %v", tt.text, tt.word, got, tt.want)
			}
		})
	}
}

// ===== MaskEmail 测试 =====

func TestMaskEmail(t *testing.T) {
	tests := []struct {
		name  string
		email string
		want  string
	}{
		{"normal email", "user@example.com", "u***@example.com"},
		{"short prefix", "a@b.com", "a***@b.com"},
		{"single char prefix", "x@test.com", "x***@test.com"},
		{"no at sign", "invalid-email", "invalid-email"},
		{"empty", "", ""},
		{"long email", "longuser@domain.org", "l***@domain.org"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MaskEmail(tt.email)
			if got != tt.want {
				t.Errorf("MaskEmail(%q) = %q; want %q", tt.email, got, tt.want)
			}
		})
	}
}

// ===== ExtractPlainText 测试 =====

func TestExtractPlainText(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		maxLen int
		want   string
	}{
		{"short text", "<p>hello</p>", 100, "hello"},
		{"needs truncation", "<p>hello world this is a test</p>", 5, "hello..."},
		{"empty input", "", 10, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ExtractPlainText(tt.input, tt.maxLen)
			if !strings.HasPrefix(got, strings.ReplaceAll(tt.want, "...", "")) {
				t.Errorf("ExtractPlainText(%q, %d) = %q; want prefix of %q", tt.input, tt.maxLen, got, tt.want)
			}
		})
	}
}

// ===== Benchmark =====

func BenchmarkStripTags(b *testing.B) {
	html := `<div><p>Hello</p><a href="https://example.com">Link</a><b>bold</b></div>`
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		StripTags(html)
	}
}

func BenchmarkEscapeHTML(b *testing.B) {
	input := `<script>alert("XSS")</script>&<>"'`
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		EscapeHTML(input)
	}
}

func BenchmarkTruncateString(b *testing.B) {
	longText := strings.Repeat("你好世界", 100)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		TruncateString(longText, 50)
	}
}

func BenchmarkMaskEmail(b *testing.B) {
	for i := 0; i < b.N; i++ {
		MaskEmail("user@example.com")
	}
}
