package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	term2svg "github.com/smford/golang-term2svg"
)

var version = "v" + term2svg.Version

func main() {
	var (
		title       = flag.String("title", "", "SVG window title (default: command name or 'Terminal')")
		columns     = flag.Int("columns", term2svg.DefaultColumns, "Terminal column width")
		rows        = flag.Int("rows", term2svg.DefaultRows, "Terminal row height")
		fontSize    = flag.Int("font-size", 14, "SVG font size in pixels")
		theme       = flag.String("theme", "dark", "Theme: 'dark', 'light', 'catppuccin', 'nord', 'frameless'")
		autoCrop    = flag.Bool("auto-crop", true, "Crop SVG canvas to visible text boundaries")
		showCursor  = flag.Bool("show-cursor", false, "Render terminal cursor")
		maskSecrets = flag.Bool("mask-secrets", false, "Redact sensitive tokens, IPs, passwords, and JWTs")
		cardHeader  = flag.String("header", "", "Subheader text inside the card")
		statusBadge = flag.String("badge", "", "Status badge text (e.g. 'ACTIVE', 'HEALTHY')")
		tabsStr     = flag.String("tabs", "", "Comma-separated tab labels")
		activeTab   = flag.Int("active-tab", 0, "Index of active tab (0-based)")
		output      = flag.String("output", "", "Output SVG file path (default: stdout)")
		showVersion = flag.Bool("version", false, "Print version and exit")
	)

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: term2svg [options] [-- command [args...]]\n\n")
		fmt.Fprintf(os.Stderr, "term2svg renders terminal commands, piped ANSI streams, or files into vector SVGs.\n\n")
		fmt.Fprintf(os.Stderr, "Examples:\n")
		fmt.Fprintf(os.Stderr, "  term2svg --title \"Directory Listing\" -- ls -la\n")
		fmt.Fprintf(os.Stderr, "  cat ansi.log | term2svg --mask-secrets > output.svg\n")
		fmt.Fprintf(os.Stderr, "  term2svg -output pods.svg -- kubectl get pods\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
	}

	flag.Parse()

	if *showVersion {
		fmt.Printf("term2svg version %s\n", version)
		os.Exit(0)
	}

	var tabs []string
	if *tabsStr != "" {
		for _, t := range strings.Split(*tabsStr, ",") {
			tTrim := strings.TrimSpace(t)
			if tTrim != "" {
				tabs = append(tabs, tTrim)
			}
		}
	}

	opts := term2svg.Options{
		Title:       *title,
		Columns:     *columns,
		Rows:        *rows,
		FontSize:    *fontSize,
		Theme:       *theme,
		AutoCrop:    *autoCrop,
		ShowCursor:  *showCursor,
		MaskSecrets: *maskSecrets,
		CardHeader:  *cardHeader,
		StatusBadge: *statusBadge,
		Tabs:        tabs,
		ActiveTab:   *activeTab,
	}

	args := flag.Args()
	var svgOutput string
	var err error

	if len(args) > 0 {
		cmdName := args[0]
		cmdArgs := args[1:]
		if opts.Title == "" {
			opts.Title = cmdName
		}
		svgOutput, err = term2svg.CommandToSVG(context.Background(), cmdName, cmdArgs, opts)
	} else {
		// Check if data is piped to stdin
		stat, _ := os.Stdin.Stat()
		if (stat.Mode() & os.ModeCharDevice) == 0 {
			if opts.Title == "" {
				opts.Title = "Terminal Stream"
			}
			svgOutput, err = term2svg.ReaderToSVG(os.Stdin, opts)
		} else {
			flag.Usage()
			os.Exit(1)
		}
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error rendering SVG: %v\n", err)
		os.Exit(1)
	}

	if *output != "" {
		if err := os.WriteFile(*output, []byte(svgOutput), 0644); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing output file %s: %v\n", *output, err)
			os.Exit(1)
		}
	} else {
		_, _ = io.WriteString(os.Stdout, svgOutput)
	}
}
