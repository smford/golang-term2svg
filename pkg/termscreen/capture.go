package termscreen

import (
	"github.com/smford/golang-term2svg/pkg/colorize"
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/creack/pty"
	"golang.org/x/term"
)

// SnapshotCurrentTerminal reads the dimensions of the current active terminal (or fallback 80x24).
// If a tmux session is active, one can capture pane output; otherwise, if stdin is a terminal,
// it determines columns and rows.
func GetCurrentTerminalSize() (cols, rows int, err error) {
	if term.IsTerminal(int(os.Stdin.Fd())) {
		cols, rows, err = term.GetSize(int(os.Stdin.Fd()))
		if err == nil && cols > 0 && rows > 0 {
			return cols, rows, nil
		}
	}
	return 80, 24, nil
}

// CaptureCommand runs a shell command in a pseudo-terminal (PTY) with specified dimensions
// and captures the resulting screen state.
func CaptureCommand(ctx context.Context, cols, rows int, command string, args ...string) (*Screen, error) {
	if cols <= 0 {
		cols = 140
	}
	if rows <= 0 {
		rows = 30
	}

	cmd := exec.CommandContext(ctx, command, args...)
	// Set terminal environment variables
	cmd.Env = append(os.Environ(),
		"TERM=xterm-256color",
		fmt.Sprintf("COLUMNS=%d", cols),
		fmt.Sprintf("LINES=%d", rows),
	)

	ptmx, err := pty.StartWithSize(cmd, &pty.Winsize{
		Rows: uint16(rows),
		Cols: uint16(cols),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to start pty: %w", err)
	}
	defer func() { _ = ptmx.Close() }()

	emu := NewEmulator(cols, rows)
	buf := make([]byte, 4096)

	done := make(chan struct{})
	go func() {
		defer close(done)
		for {
			n, readErr := ptmx.Read(buf)
			if n > 0 {
				_, _ = emu.Write(buf[:n])
			}
			if readErr != nil {
				break
			}
		}
	}()

	_ = cmd.Wait()
	// Drain remaining
	select {
	case <-done:
	case <-time.After(200 * time.Millisecond):
	}

	return emu.Screen(), nil
}

// CaptureFromReader reads ANSI stream from an io.Reader and feeds it into the emulator.
func CaptureFromReader(r io.Reader, cols, rows int) (*Screen, error) {
	emu := NewEmulator(cols, rows)
	buf := make([]byte, 4096)
	for {
		n, err := r.Read(buf)
		if n > 0 {
			_, _ = emu.Write(buf[:n])
		}
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
	}
	return emu.Screen(), nil
}

// CaptureTmuxPane captures the screen contents of a running tmux pane.
func CaptureTmuxPane(target string) (*Screen, error) {
	args := []string{"capture-pane", "-ep"}
	if target != "" {
		args = append(args, "-t", target)
	}
	out, err := exec.Command("tmux", args...).Output()
	if err != nil {
		return nil, fmt.Errorf("failed to capture tmux pane: %w", err)
	}

	dimArgs := []string{"display-message", "-p"}
	if target != "" {
		dimArgs = append(dimArgs, "-t", target)
	}
	dimArgs = append(dimArgs, "#{pane_width} #{pane_height}")
	dimOut, err := exec.Command("tmux", dimArgs...).Output()
	cols, rows := 140, 30
	if err == nil {
		_, _ = fmt.Sscanf(string(dimOut), "%d %d", &cols, &rows)
	}

	return CaptureFromReader(bytes.NewReader(out), cols, rows)
}

// CaptureLiveTerminal attempts to capture the visible screen contents of the currently active terminal emulator.
// Supported environments:
// 1. Tmux ($TMUX)
// 2. iTerm2 ($TERM_PROGRAM == "iTerm.app" or "iTerm")
// 3. Apple Terminal ($TERM_PROGRAM == "Apple_Terminal")
// 4. Kitty ($KITTY_WINDOW_ID)
// 5. WezTerm ($WEZTERM_PANE)
func CaptureLiveTerminal() (*Screen, error) {
	cols, rows, _ := GetCurrentTerminalSize()
	if cols <= 0 {
		cols = 140
	}
	if rows <= 0 {
		rows = 30
	}

	// 1. Check Tmux
	if os.Getenv("TMUX") != "" {
		return CaptureTmuxPane("")
	}

	// 2. Check iTerm2
	termProg := os.Getenv("TERM_PROGRAM")
	if termProg == "iTerm.app" || termProg == "iTerm" {
		script := `tell application "iTerm2" to tell current session of current window to get contents`
		out, err := exec.Command("osascript", "-e", script).Output()
		if err == nil && len(out) > 0 {
			lines := strings.Split(strings.TrimRight(string(out), "\r\n"), "\n")
			if len(lines) > rows {
				lines = lines[len(lines)-rows:]
			}
			text := colorize.Text(strings.Join(lines, "\n") + "\n")
			return CaptureFromReader(strings.NewReader(text), cols, rows)
		}
	}

	// 3. Check Apple Terminal
	if termProg == "Apple_Terminal" {
		script := `tell application "Terminal" to get contents of selected tab of front window`
		out, err := exec.Command("osascript", "-e", script).Output()
		if err == nil && len(out) > 0 {
			lines := strings.Split(strings.TrimRight(string(out), "\r\n"), "\n")
			if len(lines) > rows {
				lines = lines[len(lines)-rows:]
			}
			text := colorize.Text(strings.Join(lines, "\n") + "\n")
			return CaptureFromReader(strings.NewReader(text), cols, rows)
		}
	}

	// 4. Check Kitty
	if os.Getenv("KITTY_WINDOW_ID") != "" {
		out, err := exec.Command("kitty", "@", "get-text", "--extent=screen").Output()
		if err == nil && len(out) > 0 {
			return CaptureFromReader(bytes.NewReader(out), cols, rows)
		}
	}

	// 5. Check WezTerm
	if os.Getenv("WEZTERM_PANE") != "" {
		out, err := exec.Command("wezterm", "cli", "get-text", "--pane-id", os.Getenv("WEZTERM_PANE")).Output()
		if err == nil && len(out) > 0 {
			return CaptureFromReader(bytes.NewReader(out), cols, rows)
		}
	}

	return nil, fmt.Errorf("no supported live terminal environment detected (tmux, iTerm2, Terminal.app, kitty, wezterm)")
}
