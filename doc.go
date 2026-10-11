// Package term2svg provides high-level APIs to snapshot terminal commands, ANSI streams,
// tmux panes, and live terminal screens into styled vector SVG images.
//
// Key features include:
//   - Virtual terminal emulation (VT100/ANSI SGR styles, 16/256/24-bit TrueColor)
//   - Pluggable SVG rendering engine with theme support (Catppuccin, Nord, Dark, Light)
//   - Sub-pixel alignment for Unicode box-drawing glyphs
//   - Automated sensitive credential and IP redaction
//   - Automated canvas auto-cropping
package term2svg
