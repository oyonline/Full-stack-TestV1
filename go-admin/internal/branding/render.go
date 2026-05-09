package branding

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"strconv"
	"strings"
	"sync"
	"unicode"

	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
	"golang.org/x/image/math/fixed"
)

var validSizes = map[int]bool{16: true, 32: true, 64: true, 96: true, 128: true, 256: true}

// RenderDefaultLogoPNG generates a square PNG image with a colored background and a centered text character.
// text must be a single Unicode character; empty string defaults to "S".
// bgHex must be in "#RRGGBB" format.
// size must be one of {16, 32, 64, 96, 128, 256}.
func RenderDefaultLogoPNG(text, bgHex string, size int) ([]byte, error) {
	if !validSizes[size] {
		return nil, fmt.Errorf("invalid size %d: must be one of 16,32,64,96,128,256", size)
	}

	ch, err := parseTextChar(text)
	if err != nil {
		return nil, err
	}

	bg, err := parseHexColor(bgHex)
	if err != nil {
		return nil, err
	}

	cacheKey := fmt.Sprintf("%s|%s|%d", string(ch), bgHex, size)
	if cached, ok := lruGet(cacheKey); ok {
		return cached, nil
	}

	data, err := renderPNG(ch, bg, size)
	if err != nil {
		return nil, err
	}

	lruPut(cacheKey, data)
	return data, nil
}

// ETagFor returns a deterministic ETag for the given parameters plus an optional version signal.
func ETagFor(text, bgHex string, size int, brandingSig string) string {
	h := sha1.New()
	h.Write([]byte(text + "|" + bgHex + "|" + strconv.Itoa(size) + "|" + brandingSig))
	return `"` + hex.EncodeToString(h.Sum(nil))[:16] + `"`
}

func parseTextChar(text string) (rune, error) {
	runes := []rune(strings.TrimSpace(text))
	if len(runes) == 0 {
		return 'S', nil
	}
	return runes[0], nil
}

func parseHexColor(s string) (color.RGBA, error) {
	s = strings.TrimSpace(s)
	if !strings.HasPrefix(s, "#") || len(s) != 7 {
		return color.RGBA{}, fmt.Errorf("invalid hex color %q: must be #RRGGBB", s)
	}
	b, err := hex.DecodeString(s[1:])
	if err != nil || len(b) != 3 {
		return color.RGBA{}, fmt.Errorf("invalid hex color %q", s)
	}
	return color.RGBA{R: b[0], G: b[1], B: b[2], A: 255}, nil
}

func renderPNG(ch rune, bg color.RGBA, size int) ([]byte, error) {
	fontData := InterSemiBold
	if unicode.Is(unicode.Han, ch) {
		fontData = NotoSansSC
	}

	f, err := truetype.Parse(fontData)
	if err != nil {
		return nil, fmt.Errorf("parse font: %w", err)
	}

	img := image.NewRGBA(image.Rect(0, 0, size, size))
	draw.Draw(img, img.Bounds(), &image.Uniform{bg}, image.Point{}, draw.Src)

	fg := pickForeground(bg)

	fontSize := float64(size) * 0.55
	c := freetype.NewContext()
	c.SetDPI(72)
	c.SetFont(f)
	c.SetFontSize(fontSize)
	c.SetClip(img.Bounds())
	c.SetDst(img)
	c.SetSrc(&image.Uniform{fg})

	opts := truetype.Options{Size: fontSize, DPI: 72}
	face := truetype.NewFace(f, &opts)
	advance, ok := face.GlyphAdvance(ch)
	if !ok {
		advance = fixed.I(size / 2)
	}
	glyphW := int(advance >> 6)
	glyphH := int(fontSize * 0.75)

	x := (size - glyphW) / 2
	y := (size+glyphH)/2 + int(fontSize*0.05)

	pt := freetype.Pt(x, y)
	if _, err = c.DrawString(string(ch), pt); err != nil {
		return nil, fmt.Errorf("draw text: %w", err)
	}

	var buf bytes.Buffer
	if err = png.Encode(&buf, img); err != nil {
		return nil, fmt.Errorf("encode png: %w", err)
	}
	return buf.Bytes(), nil
}

// pickForeground returns white or black depending on bg luminance.
func pickForeground(bg color.RGBA) color.RGBA {
	r := float64(bg.R) / 255
	g := float64(bg.G) / 255
	b := float64(bg.B) / 255
	linearize := func(v float64) float64 {
		if v <= 0.04045 {
			return v / 12.92
		}
		return ((v + 0.055) / 1.055) * ((v + 0.055) / 1.055) // simplified pow 2.4 approximation
	}
	L := 0.2126*linearize(r) + 0.7152*linearize(g) + 0.0722*linearize(b)
	if L > 0.179 {
		return color.RGBA{R: 0, G: 0, B: 0, A: 255}
	}
	return color.RGBA{R: 255, G: 255, B: 255, A: 255}
}

// --- Simple LRU cache (256 entries) ---

type lruEntry struct {
	key  string
	data []byte
}

const lruCapacity = 256

var (
	lruMu    sync.Mutex
	lruCache = make([]lruEntry, 0, lruCapacity)
)

func lruGet(key string) ([]byte, bool) {
	lruMu.Lock()
	defer lruMu.Unlock()
	for i, e := range lruCache {
		if e.key == key {
			lruCache = append(lruCache[:i], lruCache[i+1:]...)
			lruCache = append(lruCache, e)
			return e.data, true
		}
	}
	return nil, false
}

func lruPut(key string, data []byte) {
	lruMu.Lock()
	defer lruMu.Unlock()
	for i, e := range lruCache {
		if e.key == key {
			lruCache = append(lruCache[:i], lruCache[i+1:]...)
			lruCache = append(lruCache, lruEntry{key, data})
			return
		}
	}
	if len(lruCache) >= lruCapacity {
		lruCache = lruCache[1:]
	}
	lruCache = append(lruCache, lruEntry{key, data})
}
