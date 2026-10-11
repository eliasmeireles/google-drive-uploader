package auth

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsTerminal(t *testing.T) {
	t.Run("given a regular file then it is not a terminal", func(t *testing.T) {
		f, err := os.Create(filepath.Join(t.TempDir(), "stdin"))
		require.NoError(t, err)
		defer f.Close()

		assert.False(t, isTerminal(f))
	})

	t.Run("given a closed file then it is not a terminal", func(t *testing.T) {
		f, err := os.Create(filepath.Join(t.TempDir(), "closed"))
		require.NoError(t, err)
		require.NoError(t, f.Close())

		assert.False(t, isTerminal(f))
	})

	t.Run("given /dev/null then it is not a terminal", func(t *testing.T) {
		f, err := os.Open(os.DevNull)
		require.NoError(t, err)
		defer f.Close()

		assert.False(t, isTerminal(f))
	})

	t.Run("given a pipe then it is not a terminal", func(t *testing.T) {
		r, w, err := os.Pipe()
		require.NoError(t, err)
		defer r.Close()
		defer w.Close()

		assert.False(t, isTerminal(r))
	})
}
