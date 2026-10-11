package term2svg

import (
	"testing"
)

func TestVersion(t *testing.T) {
	if Version == "" {
		t.Fatal("expected Version to not be empty")
	}
}
