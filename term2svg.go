// Package termtosvg provides high-level APIs to snapshot terminal commands, ANSI streams,
// tmux panes, and current terminal screens into styled vector SVG images.
package term2svg

import (
	"strings"
	"context"
	"io"
	"os"

	"github.com/smford/golang-term2svg/pkg/svgrender"
	"github.com/smford/golang-term2svg/pkg/termscreen"
	"github.com/smford/golang-term2svg/pkg/redact"
)

const (
	// DefaultColumns is the standard width matching modern wide terminal displays and template cards.
	DefaultColumns = 140
	// DefaultRows is the standard height for screen snapshots.
	DefaultRows = 30
)

// Options specifies high-level options for terminal-to-SVG conversion.
type Options struct {
	Title       string
	Columns     int
	Rows        int
	FontSize    int
	Theme       string   // "dark" (default), "light", "catppuccin", "nord", "frameless"
	AutoCrop    bool     // Crop canvas down to actual content dimensions
	ShowCursor  bool     // Render terminal cursor
	Tabs        []string // Navigation tab bar items
	ActiveTab   int      // Index of active tab (0-based)
	CardHeader  string   // Subheader title inside the inner card
	StatusBadge string   // Badge text (e.g. "FOLLOW: ON", "ACTIVE")
	MaskSecrets bool     // Auto-redact IPs, API tokens, passwords, and JWTs
	CustomTmpl  string   // Optional raw SVG Go template
}

// ScreenToSVG renders an existing termscreen.Screen model into SVG.
func ScreenToSVG(screen *termscreen.Screen, opts Options) (string, error) {
	cfg := svgrender.DefaultConfig()
	if opts.Title != "" {
		cfg.Title = opts.Title
	}
	if opts.Theme != "" {
		cfg.ThemeName = opts.Theme
	}
	if opts.FontSize > 0 {
		cfg.FontSize = opts.FontSize
	}
	cfg.AutoCrop = opts.AutoCrop
	cfg.ShowCursor = opts.ShowCursor
	cfg.Tabs = opts.Tabs
	cfg.ActiveTab = opts.ActiveTab
	cfg.CardHeader = opts.CardHeader
	cfg.StatusBadge = opts.StatusBadge
	if opts.CustomTmpl != "" {
		cfg.CustomTmpl = opts.CustomTmpl
	}

	renderer, err := svgrender.NewRenderer(cfg)
	if err != nil {
		return "", err
	}

	return renderer.RenderToString(screen)
}

// CommandToSVG executes a command in a virtual pseudo-terminal (PTY) and renders its output to SVG.
func CommandToSVG(ctx context.Context, cmd string, args []string, opts Options) (string, error) {
	cols := opts.Columns
	rows := opts.Rows
	if cols <= 0 {
		cols = DefaultColumns
	}
	if rows <= 0 {
		rows = DefaultRows
	}

	screen, err := termscreen.CaptureCommand(ctx, cols, rows, cmd, args...)
	if err != nil {
		return "", err
	}

	if opts.Title == "" {
		opts.Title = cmd
	}

	return ScreenToSVG(screen, opts)
}

// ReaderToSVG converts ANSI terminal output stream into SVG.
func ReaderToSVG(r io.Reader, opts Options) (string, error) {
	if opts.MaskSecrets {
		raw, err := io.ReadAll(r)
		if err != nil {
			return "", err
		}
		redactor := redact.New()
		sanitized := redactor.Apply(string(raw))
		r = strings.NewReader(sanitized)
	}
	cols := opts.Columns
	rows := opts.Rows
	if cols <= 0 {
		cols = DefaultColumns
	}
	if rows <= 0 {
		rows = DefaultRows
	}

	screen, err := termscreen.CaptureFromReader(r, cols, rows)
	if err != nil {
		return "", err
	}

	return ScreenToSVG(screen, opts)
}

// FileToSVG reads an ANSI file and converts it into SVG.
func FileToSVG(filePath string, opts Options) (string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer f.Close()
	return ReaderToSVG(f, opts)
}

// TmuxToSVG snapshots a live tmux pane and renders it to SVG.
func TmuxToSVG(paneTarget string, opts Options) (string, error) {
	screen, err := termscreen.CaptureTmuxPane(paneTarget)
	if err != nil {
		return "", err
	}
	return ScreenToSVG(screen, opts)
}

// LiveTerminalToSVG captures the visible screen of the currently active terminal session and renders it to SVG.
func LiveTerminalToSVG(opts Options) (string, error) {
	screen, err := termscreen.CaptureLiveTerminal()
	if err != nil {
		return "", err
	}
	if opts.Title == "" {
		opts.Title = "Live Terminal"
	}
	return ScreenToSVG(screen, opts)
}
