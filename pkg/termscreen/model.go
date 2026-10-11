package termscreen

// Color represents an ANSI color or RGB color.
type Color struct {
	IsSet   bool
	Is256   bool
	IsRGB   bool
	Index   uint8
	R, G, B uint8
}

// Hex returns a CSS hex color code or empty string if not set.
func (c Color) Hex() string {
	if !c.IsSet {
		return ""
	}
	if c.IsRGB {
		return rgbToHex(c.R, c.G, c.B)
	}
	if c.Is256 {
		return xterm256ToHex(c.Index)
	}
	return ansi16ToHex(c.Index)
}

func rgbToHex(r, g, b uint8) string {
	const hexChars = "0123456789abcdef"
	return string([]byte{'#', hexChars[r>>4], hexChars[r&0x0f], hexChars[g>>4], hexChars[g&0x0f], hexChars[b>>4], hexChars[b&0x0f]})
}

func ansi16ToHex(idx uint8) string {
	// Standard palette matching modern terminal & tui-logs styling
	palette := [16]string{
		"#0f172a", // 0: Black
		"#ef4444", // 1: Red
		"#10b981", // 2: Green
		"#f59e0b", // 3: Yellow
		"#3b82f6", // 4: Blue
		"#d946ef", // 5: Magenta
		"#38bdf8", // 6: Cyan
		"#cbd5e1", // 7: White
		"#64748b", // 8: Bright Black (dark dim)
		"#f87171", // 9: Bright Red
		"#34d399", // 10: Bright Green
		"#fbbf24", // 11: Bright Yellow
		"#60a5fa", // 12: Bright Blue
		"#e879f9", // 13: Bright Magenta
		"#7dd3fc", // 14: Bright Cyan
		"#ffffff", // 15: Bright White
	}
	if int(idx) < len(palette) {
		return palette[idx]
	}
	return "#e2e8f0"
}

func xterm256ToHex(idx uint8) string {
	if idx < 16 {
		return ansi16ToHex(idx)
	}
	if idx >= 16 && idx <= 231 {
		// 6x6x6 cube
		i := idx - 16
		b := i % 6
		g := (i / 6) % 6
		r := i / 36
		cv := func(v uint8) uint8 {
			if v == 0 {
				return 0
			}
			return 55 + v*40
		}
		return rgbToHex(cv(r), cv(g), cv(b))
	}
	// 232..255 grayscale ramp
	gray := 8 + (idx-232)*10
	return rgbToHex(gray, gray, gray)
}

// CellStyle represents the visual attributes of a cell.
type CellStyle struct {
	Fg        Color
	Bg        Color
	Bold      bool
	Dim       bool
	Italic    bool
	Underline bool
	Inverse   bool
}

// Cell represents an individual screen character cell.
type Cell struct {
	Char  rune
	Style CellStyle
	Width int    // display width (usually 1, 0 for trailing part of wide char)
	Link  string // OSC 8 Hyperlink target URL (if any)
}

// Line is a row of cells.
type Line []Cell

// Screen is a 2D grid representation of the terminal screen.
type Screen struct {
	Cols    int
	Rows    int
	CursorX int
	CursorY int
	Lines   []Line
}

// NewScreen creates a blank screen with given cols and rows.
func NewScreen(cols, rows int) *Screen {
	s := &Screen{
		Cols:  cols,
		Rows:  rows,
		Lines: make([]Line, rows),
	}
	for r := 0; r < rows; r++ {
		s.Lines[r] = make(Line, cols)
		for c := 0; c < cols; c++ {
			s.Lines[r][c] = Cell{Char: ' ', Width: 1}
		}
	}
	return s
}
