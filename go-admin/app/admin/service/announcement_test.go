package service

import (
	"strings"
	"testing"
)

func TestSanitizeContent_StripsScriptAndOnAttr(t *testing.T) {
	cases := []struct {
		name     string
		input    string
		mustHave []string
		mustMiss []string
	}{
		{
			name:     "strips script tag",
			input:    `<p>hello</p><script>alert(1)</script>`,
			mustHave: []string{"<p>hello</p>"},
			mustMiss: []string{"<script", "alert("},
		},
		{
			name:     "strips on* attributes",
			input:    `<a href="http://example.com" onclick="hack()">x</a>`,
			mustMiss: []string{"onclick"},
			mustHave: []string{"example.com"},
		},
		{
			name:     "preserves img src for embed",
			input:    `<p>正文</p><img src="https://cdn.example.com/a.png" alt="封面"/>`,
			mustHave: []string{"<img", "src=", "cdn.example.com"},
		},
		{
			name:     "preserves headings and lists",
			input:    `<h2>标题</h2><ul><li>a</li><li>b</li></ul>`,
			mustHave: []string{"<h2>标题</h2>", "<ul>", "<li>a</li>", "<li>b</li>"},
		},
		{
			name:     "strips iframe",
			input:    `<iframe src="http://attacker"></iframe><p>safe</p>`,
			mustMiss: []string{"<iframe"},
			mustHave: []string{"<p>safe</p>"},
		},
		{
			name:  "empty stays empty",
			input: "",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			out := SanitizeContent(tc.input)
			for _, s := range tc.mustHave {
				if !strings.Contains(out, s) {
					t.Errorf("expected %q in output, got %q", s, out)
				}
			}
			for _, s := range tc.mustMiss {
				if strings.Contains(out, s) {
					t.Errorf("did NOT expect %q in output, got %q", s, out)
				}
			}
		})
	}
}
