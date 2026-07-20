package cmd

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/canonical/gherkinator/internal/version"
)

// captureStdout routes cobra's output writer into a buffer for the
// duration of fn. The previous output writer is restored afterwards so
// other tests are not affected.
func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	var buf bytes.Buffer
	prevOut := rootCmd.OutOrStdout()
	prevErr := rootCmd.ErrOrStderr()
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)
	t.Cleanup(func() {
		rootCmd.SetOut(prevOut)
		rootCmd.SetErr(prevErr)
	})
	fn()
	return buf.String()
}

func TestVersionCommand_PrintsVersion(t *testing.T) {
	resetFlags()
	out := captureStdout(t, func() {
		rootCmd.SetArgs([]string{"version"})
		err := rootCmd.Execute()
		require.NoError(t, err)
	})
	assert.Equal(t, version.Version+"\n", out)
}

func TestVersionCommand_IsRegisteredOnRoot(t *testing.T) {
	resetFlags()
	cmds := rootCmd.Commands()
	var found bool
	for _, c := range cmds {
		if c.Name() == "version" {
			found = true
			break
		}
	}
	assert.True(t, found, "expected `version` subcommand to be registered on rootCmd")
}

func TestVersionFlag_PrintsVersion(t *testing.T) {
	resetFlags()
	out := captureStdout(t, func() {
		rootCmd.SetArgs([]string{"--version"})
		err := rootCmd.Execute()
		require.NoError(t, err)
	})
	assert.Equal(t, version.Version+"\n", out)
}

func TestVersionFlag_HasNoShorthandV(t *testing.T) {
	// `-v` must NOT be accepted as a shorthand for --version; it is
	// reserved for a future --verbose flag.
	resetFlags()
	rootCmd.SetArgs([]string{"-v"})
	err := rootCmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown shorthand flag")
}
