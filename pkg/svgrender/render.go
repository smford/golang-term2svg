package svgrender

import (
	"bytes"
	"fmt"
	"html"
	"html/template"
	"io"
	"strings"

	"github.com/smford/golang-term2svg/pkg/svgrender/templates"
	"github.com/smford/golang-term2svg/pkg/termscreen"
	"github.com/mattn/go-runewidth"
)

// TabItem represents a top navigation tab.
type TabItem struct {
	Label    string
	IsActive bool
	X        int
	Width    int
}

// Config configures SVG rendering options.
type Config struct {
	Title        string
	FontSize     int     // Default: 13
	LineHeight   int     // Default: 21
	CharWidth    float64 // Default: 8.0
	PaddingX     int     // Default: 100
	PaddingY     int     // Default: 140
	ThemeName    string  // "dark", "light", "catppuccin", "nord", "frameless"
	AutoCrop     bool    // Shrink canvas to visible content
	ShowCursor   bool    // Render blinking terminal cursor
	Tabs         []string
	ActiveTab    int
	CardHeader   string // E.g. "System Event Log (Ring Buffer - Last 250 Records)"
	StatusBadge  string // E.g. "FOLLOW: AUTO-SCROLL (ON)"
	CustomTmpl   string // Optional raw SVG Go template
}

// DefaultConfig returns production-ready defaults matching tui-logs.svg.
func DefaultConfig() Config {
	return Config{
		Title:       "Terminal",
		FontSize:    13,
		LineHeight:  21,
		CharWidth:   8.0,
		PaddingX:    100,
		PaddingY:    140,
		ThemeName:   "dark",
		AutoCrop:    false,
		ShowCursor:  false,
	}
}

// Renderer transforms a termscreen.Screen into an SVG document.
type Renderer struct {
	cfg   Config
	theme Theme
	tmpl  *template.Template
}

// NewRenderer creates a new renderer with optional configuration.
func NewRenderer(cfg Config) (*Renderer, error) {
	if cfg.FontSize <= 0 {
		cfg.FontSize = 13
	}
	if cfg.LineHeight <= 0 {
		cfg.LineHeight = 21
	}
	if cfg.CharWidth <= 0 {
		cfg.CharWidth = 8.0
	}
	if cfg.PaddingX <= 0 {
		cfg.PaddingX = 100
	}
	if cfg.PaddingY <= 0 {
		cfg.PaddingY = 140
	}

	tmplContent := templates.WindowTemplate
	if cfg.CustomTmpl != "" {
		tmplContent = cfg.CustomTmpl
	}

	funcMap := template.FuncMap{
		"sub": func(a, b int) int { return a - b },
		"add": func(a, b int) int { return a + b },
		"div": func(a, b int) int {
			if b == 0 {
				return 0
			}
			return a / b
		},
		"safeHTML": func(s string) template.HTML {
			return template.HTML(s)
		},
	}

	t, err := template.New("svg").Funcs(funcMap).Parse(tmplContent)
	if err != nil {
		return nil, fmt.Errorf("failed to parse svg template: %w", err)
	}

	theme := GetTheme(cfg.ThemeName)

	return &Renderer{
		cfg:   cfg,
		theme: theme,
		tmpl:  t,
	}, nil
}

type templateData struct {
	Width         int
	Height        int
	FontSize      int
	Title         string
	Theme         Theme
	Tabs          []TabItem
	CardY         int
	CardHeight    int
	CardHeader    string
	StatusBadge   string
	BadgeX        int
	BadgeWidth    int
	ContentY      int
	RenderedLines []string
	ShowCursor    bool
	CursorPixelX  int
	CursorPixelY  int
}

// Render writes the SVG of the screen to the provided writer.
func (r *Renderer) Render(s *termscreen.Screen, w io.Writer) error {
	renderedLines, maxCol, maxRow := r.renderScreenLines(s)

	effectiveCols := s.Cols
	effectiveRows := s.Rows

	if r.cfg.AutoCrop {
		if maxCol > 0 {
			effectiveCols = maxCol + 2
		} else {
			effectiveCols = 20
		}
		if maxRow > 0 {
			effectiveRows = maxRow + 1
		} else {
			effectiveRows = 2
		}
	}

	calcWidth := int(float64(effectiveCols)*r.cfg.CharWidth) + r.cfg.PaddingX
	if calcWidth < 650 {
		calcWidth = 650
	}

	// Calculate layout vertical offsets
	cardY := 68
	if len(r.cfg.Tabs) > 0 {
		cardY = 106
	}

	contentY := 24
	extraHeaderHeight := 0
	if r.cfg.CardHeader != "" || r.cfg.StatusBadge != "" {
		contentY = 56
		extraHeaderHeight = 36
	}

	contentHeight := (effectiveRows * r.cfg.LineHeight)
	cardHeight := contentHeight + contentY + 30
	calcHeight := cardY + cardHeight + 40

	if calcHeight < 280 {
		calcHeight = 280
	}

	// Layout Tabs
	var tabItems []TabItem
	currentTabX := 0
	for idx, tabName := range r.cfg.Tabs {
		tabWidth := int(float64(len(tabName)+6)*8.0) + 20
		if tabWidth < 110 {
			tabWidth = 110
		}
		tabItems = append(tabItems, TabItem{
			Label:    tabName,
			IsActive: idx == r.cfg.ActiveTab,
			X:        currentTabX,
			Width:    tabWidth,
		})
		currentTabX += tabWidth + 10
	}

	// Badge positioning
	badgeWidth := len(r.cfg.StatusBadge)*8 + 24
	badgeX := calcWidth - 70 - badgeWidth - 30
	if badgeX < 350 {
		badgeX = 350
	}

	cursorPixelX := int(float64(s.CursorX) * r.cfg.CharWidth)
	cursorPixelY := s.CursorY*r.cfg.LineHeight - (r.cfg.FontSize - 2)

	data := templateData{
		Width:         calcWidth,
		Height:        calcHeight,
		FontSize:      r.cfg.FontSize,
		Title:         r.cfg.Title,
		Theme:         r.theme,
		Tabs:          tabItems,
		CardY:         cardY,
		CardHeight:    cardHeight,
		CardHeader:    r.cfg.CardHeader,
		StatusBadge:   r.cfg.StatusBadge,
		BadgeX:        badgeX,
		BadgeWidth:    badgeWidth,
		ContentY:      contentY,
		RenderedLines: renderedLines,
		ShowCursor:    r.cfg.ShowCursor,
		CursorPixelX:  cursorPixelX,
		CursorPixelY:  cursorPixelY,
	}

	_ = extraHeaderHeight

	return r.tmpl.Execute(w, data)
}

// RenderToString returns the rendered SVG as a string.
func (r *Renderer) RenderToString(s *termscreen.Screen) (string, error) {
	var buf bytes.Buffer
	if err := r.Render(s, &buf); err != nil {
		return "", err
	}
	return buf.String(), nil
}

type textSpan struct {
	startCol int
	text     string
	style    termscreen.CellStyle
	link     string
}

func (r *Renderer) renderScreenLines(s *termscreen.Screen) ([]string, int, int) {
	var lines []string
	maxColFound := 0
	maxRowFound := 0

	for rowIdx, line := range s.Lines {
		yPos := rowIdx * r.cfg.LineHeight

		// Trim trailing blank cells
		lastNonSpace := -1
		for colIdx := len(line) - 1; colIdx >= 0; colIdx-- {
			if line[colIdx].Char != ' ' && line[colIdx].Char != 0 {
				lastNonSpace = colIdx
				break
			}
		}

		if lastNonSpace == -1 {
			continue
		}

		if lastNonSpace > maxColFound {
			maxColFound = lastNonSpace
		}
		maxRowFound = rowIdx

		// Coalesce adjacent characters, tracking starting grid column
		// Box-drawing glyphs (e.g. │, ╭, ╮, ╰, ╯, ─) or characters following whitespace runs
		// get their own isolated spans with explicit grid-aligned x positions to prevent sub-pixel drift.
		var spans []textSpan
		for colIdx := 0; colIdx <= lastNonSpace; colIdx++ {
			cell := line[colIdx]
			ch := cell.Char
			if ch == 0 {
				continue
			}

			isBox := isBoxDrawing(ch)
			prevIsBox := false
			if len(spans) > 0 && len(spans[len(spans)-1].text) > 0 {
				prevRunes := []rune(spans[len(spans)-1].text)
				prevIsBox = isBoxDrawing(prevRunes[len(prevRunes)-1])
			}

			// Do not coalesce across box-drawing boundaries or between vertical box borders
			canCoalesce := len(spans) > 0 &&
				spans[len(spans)-1].style == cell.Style &&
				spans[len(spans)-1].link == cell.Link &&
				!isBox && !prevIsBox

			if canCoalesce {
				spans[len(spans)-1].text += string(ch)
			} else {
				spans = append(spans, textSpan{
					startCol: colIdx,
					text:     string(ch),
					style:    cell.Style,
					link:     cell.Link,
				})
			}
		}

		var lineBuilder strings.Builder
		for _, span := range spans {
			bgHex := span.style.Bg.Hex()
			if span.style.Inverse {
				bgHex = span.style.Fg.Hex()
				if bgHex == "" {
					bgHex = r.theme.DefaultText
				}
			}
			if bgHex != "" {
				xPos := float64(span.startCol) * r.cfg.CharWidth
				spanWidth := float64(ComputeDisplayWidth(span.text)) * r.cfg.CharWidth
				yBg := yPos - r.cfg.FontSize + 3
				lineBuilder.WriteString(fmt.Sprintf(`<rect x="%.2f" y="%d" width="%.2f" height="%d" fill="%s" rx="2"/>`, xPos, yBg, spanWidth, r.cfg.LineHeight, bgHex))
			}
		}
		lineBuilder.WriteString(fmt.Sprintf(`<text x="0" y="%d" class="tui">`, yPos))

		for _, span := range spans {
			escapedText := html.EscapeString(span.text)
			escapedText = strings.ReplaceAll(escapedText, " ", "&#160;")

			xPos := float64(span.startCol) * r.cfg.CharWidth
			attrs := buildSpanAttributes(span.style, r.theme)

			spanXML := ""
			if attrs != "" {
				spanXML = fmt.Sprintf(`<tspan x="%.2f" %s>%s</tspan>`, xPos, attrs, escapedText)
			} else {
				spanXML = fmt.Sprintf(`<tspan x="%.2f">%s</tspan>`, xPos, escapedText)
			}

			if span.link != "" {
				lineBuilder.WriteString(fmt.Sprintf(`<a href="%s" target="_blank">%s</a>`, html.EscapeString(span.link), spanXML))
			} else {
				lineBuilder.WriteString(spanXML)
			}
		}

		lineBuilder.WriteString("</text>")
		lines = append(lines, lineBuilder.String())
	}

	return lines, maxColFound, maxRowFound
}

func buildSpanAttributes(style termscreen.CellStyle, th Theme) string {
	var classes []string
	var styles []string

	if style.Bold {
		classes = append(classes, "bold")
	}
	if style.Dim {
		classes = append(classes, "dim")
	}
	if style.Italic {
		classes = append(classes, "italic")
	}
	if style.Underline {
		classes = append(classes, "underline")
	}

	fgHex := style.Fg.Hex()
	bgHex := style.Bg.Hex()

	if style.Inverse {
		fgHex, bgHex = bgHex, fgHex
		if fgHex == "" {
			fgHex = th.WindowBg
		}
		if bgHex == "" {
			bgHex = th.DefaultText
		}
	}

	if fgHex != "" {
		styles = append(styles, fmt.Sprintf("fill: %s;", fgHex))
	}
	if bgHex != "" {
		styles = append(styles, fmt.Sprintf("background-color: %s;", bgHex))
	}

	var parts []string
	if len(classes) > 0 {
		parts = append(parts, fmt.Sprintf(`class="%s"`, strings.Join(classes, " ")))
	}
	if len(styles) > 0 {
		parts = append(parts, fmt.Sprintf(`style="%s"`, strings.Join(styles, " ")))
	}

	return strings.Join(parts, " ")
}

// ComputeDisplayWidth returns the visual character width using runewidth.
func ComputeDisplayWidth(s string) int {
	return runewidth.StringWidth(s)
}

func isBoxDrawing(r rune) bool {
	return r >= 0x2500 && r <= 0x257F
}
