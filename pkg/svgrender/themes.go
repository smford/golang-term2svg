package svgrender

// Theme defines colors and styling parameters for terminal window rendering.
type Theme struct {
	Name            string
	WindowBg        string
	WindowBorder    string
	HeaderBg        string
	HeaderBorder    string
	CardBg          string
	CardBorder      string
	TitleColor      string
	DefaultText     string
	ActiveTabBg     string
	ActiveTabText   string
	InactiveTabBg   string
	InactiveTabText string
	ShowControls    bool
	DropShadow      bool
}

// Built-in themes
var (
	ThemeDark = Theme{
		Name:            "dark",
		WindowBg:        "#0f172a",
		WindowBorder:    "#334155",
		HeaderBg:        "#1e293b",
		HeaderBorder:    "#334155",
		CardBg:          "#121826",
		CardBorder:      "#3d4255",
		TitleColor:      "#64748b",
		DefaultText:     "#e2e8f0",
		ActiveTabBg:     "#2563eb",
		ActiveTabText:   "#ffffff",
		InactiveTabBg:   "#1e293b",
		InactiveTabText: "#94a3b8",
		ShowControls:    true,
		DropShadow:      true,
	}

	ThemeLight = Theme{
		Name:            "light",
		WindowBg:        "#f8fafc",
		WindowBorder:    "#cbd5e1",
		HeaderBg:        "#e2e8f0",
		HeaderBorder:    "#cbd5e1",
		CardBg:          "#ffffff",
		CardBorder:      "#e2e8f0",
		TitleColor:      "#475569",
		DefaultText:     "#0f172a",
		ActiveTabBg:     "#2563eb",
		ActiveTabText:   "#ffffff",
		InactiveTabBg:   "#e2e8f0",
		InactiveTabText: "#64748b",
		ShowControls:    true,
		DropShadow:      true,
	}

	ThemeCatppuccin = Theme{
		Name:            "catppuccin",
		WindowBg:        "#1e1e2e",
		WindowBorder:    "#45475a",
		HeaderBg:        "#181825",
		HeaderBorder:    "#313244",
		CardBg:          "#11111b",
		CardBorder:      "#313244",
		TitleColor:      "#a6adc8",
		DefaultText:     "#cdd6f4",
		ActiveTabBg:     "#cba6f7",
		ActiveTabText:   "#11111b",
		InactiveTabBg:   "#181825",
		InactiveTabText: "#9399b2",
		ShowControls:    true,
		DropShadow:      true,
	}

	ThemeNord = Theme{
		Name:            "nord",
		WindowBg:        "#2e3440",
		WindowBorder:    "#4c566a",
		HeaderBg:        "#3b4252",
		HeaderBorder:    "#434c5e",
		CardBg:          "#242933",
		CardBorder:      "#434c5e",
		TitleColor:      "#d8dee9",
		DefaultText:     "#eceff4",
		ActiveTabBg:     "#88c0d0",
		ActiveTabText:   "#2e3440",
		InactiveTabBg:   "#3b4252",
		InactiveTabText: "#81a1c1",
		ShowControls:    true,
		DropShadow:      true,
	}

	ThemeFrameless = Theme{
		Name:            "frameless",
		WindowBg:        "#0f172a",
		WindowBorder:    "none",
		HeaderBg:        "transparent",
		HeaderBorder:    "none",
		CardBg:          "#0f172a",
		CardBorder:      "none",
		TitleColor:      "#64748b",
		DefaultText:     "#e2e8f0",
		ActiveTabBg:     "#2563eb",
		ActiveTabText:   "#ffffff",
		InactiveTabBg:   "#1e293b",
		InactiveTabText: "#94a3b8",
		ShowControls:    false,
		DropShadow:      false,
	}
)

// GetTheme returns the theme by name, or ThemeDark if not found.
func GetTheme(name string) Theme {
	switch name {
	case "light":
		return ThemeLight
	case "catppuccin":
		return ThemeCatppuccin
	case "nord":
		return ThemeNord
	case "frameless":
		return ThemeFrameless
	case "dark":
		fallthrough
	default:
		return ThemeDark
	}
}
