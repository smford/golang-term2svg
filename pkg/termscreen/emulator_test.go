package termscreen

import "testing"

func TestEmulatorBasic(t *testing.T) {
	emu := NewEmulator(40, 5)
	_, err := emu.Write([]byte("Hello, World!\nLine 2"))
	if err != nil {
		t.Fatalf("unexpected write error: %v", err)
	}

	screen := emu.Screen()
	if screen.Rows != 5 || screen.Cols != 40 {
		t.Fatalf("unexpected screen size: %dx%d", screen.Cols, screen.Rows)
	}

	line0 := string(screen.Lines[0][0].Char) + string(screen.Lines[0][1].Char)
	if line0 != "He" {
		t.Errorf("expected 'He', got '%s'", line0)
	}

	line1 := string(screen.Lines[1][0].Char) + string(screen.Lines[1][1].Char)
	if line1 != "Li" {
		t.Errorf("expected 'Li', got '%s'", line1)
	}
}

func TestEmulatorAnsiColors(t *testing.T) {
	emu := NewEmulator(40, 5)
	// Green text: \x1b[32mOK\x1b[0m
	_, err := emu.Write([]byte("\x1b[32mOK\x1b[0m"))
	if err != nil {
		t.Fatalf("write failed: %v", err)
	}

	screen := emu.Screen()
	cell := screen.Lines[0][0]
	if cell.Char != 'O' {
		t.Fatalf("expected 'O', got '%c'", cell.Char)
	}
	if !cell.Style.Fg.IsSet || cell.Style.Fg.Hex() != "#10b981" {
		t.Fatalf("expected green hex #10b981, got %v (%s)", cell.Style.Fg, cell.Style.Fg.Hex())
	}
}
