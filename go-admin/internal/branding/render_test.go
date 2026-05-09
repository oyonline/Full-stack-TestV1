package branding

import (
	"testing"
)

func TestRenderDefaultLogoPNG_ValidInput(t *testing.T) {
	tests := []struct {
		text  string
		bg    string
		size  int
	}{
		{"E", "#1F2937", 96},
		{"S", "#1d4ed8", 64},
		{"", "#000000", 128},
		{"系", "#1d4ed8", 96},
	}

	for _, tc := range tests {
		data, err := RenderDefaultLogoPNG(tc.text, tc.bg, tc.size)
		if err != nil {
			t.Errorf("RenderDefaultLogoPNG(%q, %q, %d) error: %v", tc.text, tc.bg, tc.size, err)
			continue
		}
		if len(data) < 100 {
			t.Errorf("expected non-trivial PNG, got %d bytes", len(data))
		}
		// PNG magic bytes
		if data[0] != 0x89 || data[1] != 'P' || data[2] != 'N' || data[3] != 'G' {
			t.Errorf("output is not a valid PNG")
		}
	}
}

func TestRenderDefaultLogoPNG_InvalidSize(t *testing.T) {
	_, err := RenderDefaultLogoPNG("E", "#000000", 100)
	if err == nil {
		t.Error("expected error for invalid size 100")
	}
}

func TestRenderDefaultLogoPNG_InvalidColor(t *testing.T) {
	_, err := RenderDefaultLogoPNG("E", "notacolor", 64)
	if err == nil {
		t.Error("expected error for invalid color")
	}
}

func TestETagFor(t *testing.T) {
	e1 := ETagFor("E", "#1F2937", 96, "v1")
	e2 := ETagFor("E", "#1F2937", 96, "v1")
	e3 := ETagFor("E", "#1F2937", 96, "v2")

	if e1 != e2 {
		t.Error("same inputs should produce same ETag")
	}
	if e1 == e3 {
		t.Error("different brandingSig should produce different ETag")
	}
}
