package term2svg

import (
	"runtime/debug"
	"strings"
)

// Version represents the current Semantic Version (SemVer) of the golang-term2svg module.
var Version = "0.1.0"

func init() {
	if info, ok := debug.ReadBuildInfo(); ok {
		for _, dep := range info.Deps {
			if dep.Path == "github.com/smford/golang-term2svg" && dep.Version != "" && dep.Version != "(devel)" {
				Version = strings.TrimPrefix(dep.Version, "v")
				return
			}
		}
		if info.Main.Path == "github.com/smford/golang-term2svg" && info.Main.Version != "" && info.Main.Version != "(devel)" {
			Version = strings.TrimPrefix(info.Main.Version, "v")
		}
	}
}
