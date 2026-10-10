package termscreen

import (
	"github.com/mattn/go-runewidth"
	"strconv"
	"strings"
)

// Emulator maintains screen state by processing ANSI / VT100 / xterm sequences.
type Emulator struct {
	screen       *Screen
	altScreen    *Screen
	isAltScreen  bool
	cursorX      int
	cursorY      int
	savedX       int
	savedY       int
	currentStyle CellStyle
	currentLink  string
}

// NewEmulator creates an emulator initialized with a new Screen.
func NewEmulator(cols, rows int) *Emulator {
	if cols <= 0 {
		cols = 80
	}
	if rows <= 0 {
		rows = 24
	}
	return &Emulator{
		screen:    NewScreen(cols, rows),
		altScreen: NewScreen(cols, rows),
	}
}

// Screen returns the current screen representation.
func (e *Emulator) Screen() *Screen {
	e.screen.CursorX = e.cursorX
	e.screen.CursorY = e.cursorY
	return e.screen
}

// Write processes raw terminal output stream and updates the internal grid.
func (e *Emulator) Write(p []byte) (n int, err error) {
	runes := []rune(string(p))
	i := 0
	length := len(runes)

	for i < length {
		r := runes[i]
		switch {
		case r == '\x1b': // ESC
			i = e.handleEscape(runes, i)
		case r == '\r':
			e.cursorX = 0
			i++
		case r == '\n':
			e.cursorX = 0
			e.cursorY++
			if e.cursorY >= e.screen.Rows {
				e.scrollUp(1)
				e.cursorY = e.screen.Rows - 1
			}
			i++
		case r == '\b':
			if e.cursorX > 0 {
				e.cursorX--
			}
			i++
		case r == '\t':
			// Tab stop (every 8 spaces)
			nextTab := ((e.cursorX / 8) + 1) * 8
			if nextTab >= e.screen.Cols {
				nextTab = e.screen.Cols - 1
			}
			for e.cursorX < nextTab {
				e.putChar(' ')
			}
			i++
		case r >= 32:
			e.putChar(r)
			i++
		default:
			i++
		}
	}
	return len(p), nil
}

func (e *Emulator) putChar(r rune) {
	w := runewidth.RuneWidth(r)
	if w <= 0 {
		w = 1
	}
	if e.cursorX+w > e.screen.Cols {
		e.cursorX = 0
		e.cursorY++
		if e.cursorY >= e.screen.Rows {
			e.scrollUp(1)
			e.cursorY = e.screen.Rows - 1
		}
	}
	if e.cursorY < e.screen.Rows && e.cursorX < e.screen.Cols {
		e.screen.Lines[e.cursorY][e.cursorX] = Cell{
			Char:  r,
			Style: e.currentStyle,
			Width: w,
			Link:  e.currentLink,
		}
		// If wide character (2 cells), occupy the second cell with zero rune
		if w > 1 && e.cursorX+1 < e.screen.Cols {
			e.screen.Lines[e.cursorY][e.cursorX+1] = Cell{
				Char:  0,
				Style: e.currentStyle,
				Width: 0,
			}
		}
		e.cursorX += w
	}
}

func (e *Emulator) scrollUp(lines int) {
	for l := 0; l < lines; l++ {
		copy(e.screen.Lines[0:], e.screen.Lines[1:])
		blankRow := make(Line, e.screen.Cols)
		for c := 0; c < e.screen.Cols; c++ {
			blankRow[c] = Cell{Char: ' ', Width: 1}
		}
		e.screen.Lines[e.screen.Rows-1] = blankRow
	}
}

func (e *Emulator) handleEscape(runes []rune, start int) int {
	if start+1 >= len(runes) {
		return start + 1
	}

	next := runes[start+1]
	switch next {
	case '[': // CSI - Control Sequence Introducer
		return e.handleCSI(runes, start+2)
	case ']': // OSC - Operating System Command
		return e.handleOSC(runes, start+2)
	case '7': // Save cursor (DECSC)
		e.savedX = e.cursorX
		e.savedY = e.cursorY
		return start + 2
	case '8': // Restore cursor (DECRC)
		e.cursorX = e.savedX
		e.cursorY = e.savedY
		return start + 2
	case 'c': // RIS (Reset to Initial State)
		*e = *NewEmulator(e.screen.Cols, e.screen.Rows)
		return start + 2
	default:
		return start + 2
	}
}

func (e *Emulator) handleOSC(runes []rune, start int) int {
	i := start
	end := -1
	for i < len(runes) {
		if runes[i] == '' {
			end = i
			i++
			break
		}
		if runes[i] == '' && i+1 < len(runes) && runes[i+1] == '\\' {
			end = i
			i += 2
			break
		}
		i++
	}

	if end > start {
		oscPayload := string(runes[start:end])
		if strings.HasPrefix(oscPayload, "8;") {
			rest := oscPayload[2:]
			semi := strings.Index(rest, ";")
			if semi != -1 {
				e.currentLink = rest[semi+1:]
			}
		}
	}
	return i
}
func (e *Emulator) handleCSI(runes []rune, start int) int {
	i := start
	for i < len(runes) {
		r := runes[i]
		if (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z') || r == '@' || r == '`' {
			// Found terminator
			cmd := r
			paramsStr := string(runes[start:i])
			e.executeCSI(cmd, paramsStr)
			return i + 1
		}
		i++
	}
	return i
}

func (e *Emulator) executeCSI(cmd rune, paramsStr string) {
	// Strip leading '?' if private mode (e.g., ?25h)
	isPrivate := false
	if strings.HasPrefix(paramsStr, "?") {
		isPrivate = true
		paramsStr = paramsStr[1:]
	}

	var params []int
	if len(paramsStr) > 0 {
		parts := strings.Split(paramsStr, ";")
		for _, p := range parts {
			val, _ := strconv.Atoi(p)
			params = append(params, val)
		}
	}

	param := func(idx, defaultVal int) int {
		if idx < len(params) && params[idx] > 0 {
			return params[idx]
		}
		return defaultVal
	}

	switch cmd {
	case 'm': // SGR - Select Graphic Rendition
		e.handleSGR(params)

	case 'A': // Cursor Up
		n := param(0, 1)
		e.cursorY -= n
		if e.cursorY < 0 {
			e.cursorY = 0
		}

	case 'B': // Cursor Down
		n := param(0, 1)
		e.cursorY += n
		if e.cursorY >= e.screen.Rows {
			e.cursorY = e.screen.Rows - 1
		}

	case 'C': // Cursor Forward
		n := param(0, 1)
		e.cursorX += n
		if e.cursorX >= e.screen.Cols {
			e.cursorX = e.screen.Cols - 1
		}

	case 'D': // Cursor Back
		n := param(0, 1)
		e.cursorX -= n
		if e.cursorX < 0 {
			e.cursorX = 0
		}

	case 'H', 'f': // Cursor Position (row, col) 1-based
		row := param(0, 1) - 1
		col := param(1, 1) - 1
		if row < 0 {
			row = 0
		}
		if row >= e.screen.Rows {
			row = e.screen.Rows - 1
		}
		if col < 0 {
			col = 0
		}
		if col >= e.screen.Cols {
			col = e.screen.Cols - 1
		}
		e.cursorY = row
		e.cursorX = col

	case 'J': // Erase in Display
		mode := param(0, 0)
		switch mode {
		case 0: // clear cursor to end
			for c := e.cursorX; c < e.screen.Cols; c++ {
				e.screen.Lines[e.cursorY][c] = Cell{Char: ' ', Width: 1}
			}
			for r := e.cursorY + 1; r < e.screen.Rows; r++ {
				for c := 0; c < e.screen.Cols; c++ {
					e.screen.Lines[r][c] = Cell{Char: ' ', Width: 1}
				}
			}
		case 1: // clear start to cursor
			for r := 0; r < e.cursorY; r++ {
				for c := 0; c < e.screen.Cols; c++ {
					e.screen.Lines[r][c] = Cell{Char: ' ', Width: 1}
				}
			}
			for c := 0; c <= e.cursorX && c < e.screen.Cols; c++ {
				e.screen.Lines[e.cursorY][c] = Cell{Char: ' ', Width: 1}
			}
		case 2, 3: // clear entire screen
			for r := 0; r < e.screen.Rows; r++ {
				for c := 0; c < e.screen.Cols; c++ {
					e.screen.Lines[r][c] = Cell{Char: ' ', Width: 1}
				}
			}
		}

	case 'K': // Erase in Line
		mode := param(0, 0)
		switch mode {
		case 0: // clear cursor to end of line
			for c := e.cursorX; c < e.screen.Cols; c++ {
				e.screen.Lines[e.cursorY][c] = Cell{Char: ' ', Width: 1}
			}
		case 1: // clear start of line to cursor
			for c := 0; c <= e.cursorX && c < e.screen.Cols; c++ {
				e.screen.Lines[e.cursorY][c] = Cell{Char: ' ', Width: 1}
			}
		case 2: // clear entire line
			for c := 0; c < e.screen.Cols; c++ {
				e.screen.Lines[e.cursorY][c] = Cell{Char: ' ', Width: 1}
			}
		}

	case 's': // Save cursor
		e.savedX = e.cursorX
		e.savedY = e.cursorY

	case 'u': // Restore cursor
		e.cursorX = e.savedX
		e.cursorY = e.savedY

	case 'h', 'l':
		if isPrivate {
			for _, p := range params {
				switch p {
				case 1049, 1047, 47: // Alternate Screen Buffer
					if cmd == 'h' {
						if !e.isAltScreen {
							e.isAltScreen = true
							e.screen, e.altScreen = e.altScreen, e.screen
						}
					} else {
						if e.isAltScreen {
							e.isAltScreen = false
							e.screen, e.altScreen = e.altScreen, e.screen
						}
					}
				}
			}
		}
	}
}

func (e *Emulator) handleSGR(params []int) {
	if len(params) == 0 {
		e.currentStyle = CellStyle{}
		return
	}

	for i := 0; i < len(params); i++ {
		p := params[i]
		switch {
		case p == 0:
			e.currentStyle = CellStyle{}
		case p == 1:
			e.currentStyle.Bold = true
		case p == 2:
			e.currentStyle.Dim = true
		case p == 3:
			e.currentStyle.Italic = true
		case p == 4:
			e.currentStyle.Underline = true
		case p == 7:
			e.currentStyle.Inverse = true
		case p == 22:
			e.currentStyle.Bold = false
			e.currentStyle.Dim = false
		case p == 23:
			e.currentStyle.Italic = false
		case p == 24:
			e.currentStyle.Underline = false
		case p == 27:
			e.currentStyle.Inverse = false

		// 30..37 Standard Foreground
		case p >= 30 && p <= 37:
			e.currentStyle.Fg = Color{IsSet: true, Index: uint8(p - 30)}
		case p == 38: // Extended Foreground
			if i+2 < len(params) && params[i+1] == 5 {
				// 256 color: 38;5;idx
				e.currentStyle.Fg = Color{IsSet: true, Is256: true, Index: uint8(params[i+2])}
				i += 2
			} else if i+4 < len(params) && params[i+1] == 2 {
				// RGB color: 38;2;r;g;b
				e.currentStyle.Fg = Color{
					IsSet: true,
					IsRGB: true,
					R:     uint8(params[i+2]),
					G:     uint8(params[i+3]),
					B:     uint8(params[i+4]),
				}
				i += 4
			}
		case p == 39: // Default Foreground
			e.currentStyle.Fg = Color{}

		// 40..47 Standard Background
		case p >= 40 && p <= 47:
			e.currentStyle.Bg = Color{IsSet: true, Index: uint8(p - 40)}
		case p == 48: // Extended Background
			if i+2 < len(params) && params[i+1] == 5 {
				e.currentStyle.Bg = Color{IsSet: true, Is256: true, Index: uint8(params[i+2])}
				i += 2
			} else if i+4 < len(params) && params[i+1] == 2 {
				e.currentStyle.Bg = Color{
					IsSet: true,
					IsRGB: true,
					R:     uint8(params[i+2]),
					G:     uint8(params[i+3]),
					B:     uint8(params[i+4]),
				}
				i += 4
			}
		case p == 49: // Default Background
			e.currentStyle.Bg = Color{}

		// 90..97 Bright Foreground
		case p >= 90 && p <= 97:
			e.currentStyle.Fg = Color{IsSet: true, Index: uint8(p - 90 + 8)}

		// 100..107 Bright Background
		case p >= 100 && p <= 107:
			e.currentStyle.Bg = Color{IsSet: true, Index: uint8(p - 100 + 8)}
		}
	}
}
