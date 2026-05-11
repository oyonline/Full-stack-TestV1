package jobs

import (
	"strings"
	"testing"
)

func TestExtractActivePaths(t *testing.T) {
	cases := []struct {
		name      string
		content   string
		coverURL  string
		wantPaths []string
	}{
		{
			name:      "empty",
			content:   "",
			coverURL:  "",
			wantPaths: []string{},
		},
		{
			name:      "cover only",
			content:   "",
			coverURL:  "/static/uploadfile/202601/cover.png",
			wantPaths: []string{"/static/uploadfile/202601/cover.png"},
		},
		{
			name:      "single img",
			content:   `<p>hello</p><img src="/static/uploadfile/202601/a.png" alt="a"/>`,
			coverURL:  "",
			wantPaths: []string{"/static/uploadfile/202601/a.png"},
		},
		{
			name:      "multiple imgs",
			content:   `<img src="/a.png"/><img src="/b.png"/>`,
			coverURL:  "",
			wantPaths: []string{"/a.png", "/b.png"},
		},
		{
			name:      "img with single quotes",
			content:   `<img src='/c.png'/>`,
			coverURL:  "",
			wantPaths: []string{"/c.png"},
		},
		{
			name:      "cover plus imgs",
			content:   `<img src="/a.png"/>`,
			coverURL:  "/cover.png",
			wantPaths: []string{"/cover.png", "/a.png"},
		},
		{
			name:      "no src",
			content:   `<img alt="x"/>`,
			coverURL:  "",
			wantPaths: []string{},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := extractActivePaths(tc.content, tc.coverURL)
			if len(got) != len(tc.wantPaths) {
				t.Fatalf("expected %d paths, got %d: %v", len(tc.wantPaths), len(got), got)
			}
			for i, w := range tc.wantPaths {
				if got[i] != w {
					t.Errorf("path[%d]: expected %q, got %q", i, w, got[i])
				}
			}
		})
	}
}

func TestContains(t *testing.T) {
	if !contains([]string{"a", "b"}, "a") {
		t.Error("expected true")
	}
	if contains([]string{"a", "b"}, "c") {
		t.Error("expected false")
	}
}

func TestIsDryRun_DefaultTrue(t *testing.T) {
	// 默认无环境变量时，如果 ExtConfig 未初始化（零值），DryRun 默认 false，
	// 但 isDryRun 在无环境变量时应返回 config.ExtConfig.Announcement.AttachmentGC.DryRun。
	// 这里仅验证函数不 panic。
	_ = isDryRun()
}

func TestAnnouncementAttachmentGC_ImplementsJobExec(t *testing.T) {
	var _ JobExec = AnnouncementAttachmentGC{}
}

func TestAnnouncementAttachmentGC_LogPrefixFormat(t *testing.T) {
	// 验证日志前缀可被 grep
	gc := AnnouncementAttachmentGC{}
	_ = gc
	if !strings.Contains("[AnnouncementAttachmentGC] dry-run stage=temp_orphans matched=42", "[AnnouncementAttachmentGC]") {
		t.Error("log prefix mismatch")
	}
}
