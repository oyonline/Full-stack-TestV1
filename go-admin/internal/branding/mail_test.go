package branding

import (
	"os"
	"strings"
	"testing"
)

func TestRenderEmailTemplate_ContainsBase64Logo(t *testing.T) {
	html, err := RenderEmailTemplate(
		"SystemName",
		"欢迎使用",
		"<p>这是一封测试邮件。</p>",
		"© 2025 SystemName. All rights reserved.",
		"#1d4ed8",
	)
	if err != nil {
		t.Fatalf("RenderEmailTemplate error: %v", err)
	}

	if !strings.Contains(html, "data:image/png;base64,") {
		t.Fatal("email HTML does not contain base64 PNG logo")
	}

	// Write to /tmp/test-mail.html for manual inspection and e2e verification.
	if err = os.WriteFile("/tmp/test-mail.html", []byte(html), 0644); err != nil {
		t.Fatalf("write /tmp/test-mail.html: %v", err)
	}
	t.Log("wrote /tmp/test-mail.html")
}

func TestEmailLogoBase64_NoChinese(t *testing.T) {
	uri, err := EmailLogoBase64("E", "#1F2937")
	if err != nil {
		t.Fatalf("EmailLogoBase64 error: %v", err)
	}
	if !strings.HasPrefix(uri, "data:image/png;base64,") {
		t.Errorf("expected data URI prefix, got: %s", uri[:min(30, len(uri))])
	}
}

func TestEmailLogoBase64_Chinese(t *testing.T) {
	uri, err := EmailLogoBase64("系", "#1d4ed8")
	if err != nil {
		t.Fatalf("EmailLogoBase64 error: %v", err)
	}
	if !strings.HasPrefix(uri, "data:image/png;base64,") {
		t.Errorf("expected data URI prefix, got: %s", uri[:min(30, len(uri))])
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
