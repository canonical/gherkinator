package version

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestVersion_DefaultIsDev verifies that the Version variable retains its
// default "dev" sentinel when the binary is built without -ldflags
// injection (e.g. when `go test` compiles the package).
func TestVersion_DefaultIsDev(t *testing.T) {
	assert.Equal(t, "dev", Version)
}
