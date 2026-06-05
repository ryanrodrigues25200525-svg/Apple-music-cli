package art

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRenderPNGAsANSIHalfBlocks(t *testing.T) {
	path := writeTestPNG(t)

	got, err := Render(path, 2, 1)
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}

	if strings.Count(got, "▀") != 2 {
		t.Fatalf("Render() block count = %d, want 2; output %q", strings.Count(got, "▀"), got)
	}
	if !strings.Contains(got, "\x1b[38;2;") || !strings.Contains(got, "\x1b[48;2;") {
		t.Fatalf("Render() missing truecolor escape sequences: %q", got)
	}
}

func TestRenderMissingFile(t *testing.T) {
	if _, err := Render(filepath.Join(t.TempDir(), "missing.png"), 2, 1); err == nil {
		t.Fatal("Render() error = nil, want missing-file error")
	}
}

func TestRenderInvalidSize(t *testing.T) {
	path := writeTestPNG(t)
	if _, err := Render(path, 0, 1); err == nil {
		t.Fatal("Render() error = nil, want invalid-size error")
	}
}

func TestResizeBilinearDimensions(t *testing.T) {
	img := image.NewNRGBA(image.Rect(0, 0, 4, 4))
	resized := resizeBilinear(img, 3, 2)
	if got := resized.Bounds(); got.Dx() != 3 || got.Dy() != 2 {
		t.Fatalf("resizeBilinear() bounds = %v, want 3x2", got)
	}
}

func writeTestPNG(t *testing.T) string {
	t.Helper()

	img := image.NewNRGBA(image.Rect(0, 0, 2, 2))
	img.Set(0, 0, color.NRGBA{R: 255, A: 255})
	img.Set(1, 0, color.NRGBA{G: 255, A: 255})
	img.Set(0, 1, color.NRGBA{B: 255, A: 255})
	img.Set(1, 1, color.NRGBA{R: 255, G: 255, A: 255})

	path := filepath.Join(t.TempDir(), "art.png")
	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("create test png: %v", err)
	}
	defer f.Close()

	if err := png.Encode(f, img); err != nil {
		t.Fatalf("encode test png: %v", err)
	}
	return path
}
