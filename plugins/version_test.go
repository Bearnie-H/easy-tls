package plugins

import (
	"fmt"
	"strings"
	"testing"
)

// TestVersionParse is the
func TestVersionParse(T *testing.T) {

	var n, v string

	m, err := fmt.Sscanf(strings.Split("connectivity-1.3.242 (client)", " ")[0], "%s-%s", &n, &v)
	if err != nil {
		T.FailNow()
	}
	if m == 0 {
		T.FailNow()
	}

	_, err = ParseVersion(v)
	if err != nil {
		T.FailNow()
	}
}
