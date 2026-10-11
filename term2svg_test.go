package term2svg

import (
	"strings"
	"testing"
)

func TestReaderToSVG(t *testing.T) {
	input := "Hello World\n"
	r := strings.NewReader(input)

	svg, err := ReaderToSVG(r, Options{
		Title:    "Test Window",
		Theme:    "dark",
		AutoCrop: true,
	})
	if err != nil {
		t.Fatalf("unexpected error rendering SVG: %v", err)
	}

	if !strings.Contains(svg, "<svg") {
		t.Errorf("expected SVG output to contain '<svg', got %s", svg)
	}
	// SVG XML-escapes spaces as &#160;
	if !strings.Contains(svg, "Hello") || !strings.Contains(svg, "World") {
		t.Errorf("expected SVG output to contain 'Hello' and 'World', got %s", svg)
	}
}
