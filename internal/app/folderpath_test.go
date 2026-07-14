package app

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeFolderCreator struct {
	created []string
	failOn  string
}

func (f *fakeFolderCreator) FindOrCreateFolder(_ context.Context, name string, parentID string) (string, error) {
	if name == f.failOn {
		return "", errors.New("drive error")
	}
	f.created = append(f.created, name)
	return parentID + "/" + name, nil
}

func TestResolveFolderPath(t *testing.T) {
	t.Run("given a single segment then resolves one folder under the root", func(t *testing.T) {
		fc := &fakeFolderCreator{}
		id, err := ResolveFolderPath(context.Background(), fc, "MONGODB", "root")
		require.NoError(t, err)
		assert.Equal(t, "root/MONGODB", id)
		assert.Equal(t, []string{"MONGODB"}, fc.created)
	})

	t.Run("given a nested path then resolves each segment under the previous one", func(t *testing.T) {
		fc := &fakeFolderCreator{}
		id, err := ResolveFolderPath(context.Background(), fc, "ORACLE-VPS-01/MONGODB", "root")
		require.NoError(t, err)
		assert.Equal(t, "root/ORACLE-VPS-01/MONGODB", id)
		assert.Equal(t, []string{"ORACLE-VPS-01", "MONGODB"}, fc.created)
	})

	t.Run("given blank and duplicate separators then ignores empty segments", func(t *testing.T) {
		fc := &fakeFolderCreator{}
		id, err := ResolveFolderPath(context.Background(), fc, "/ORACLE-VPS-01//MONGODB/", "root")
		require.NoError(t, err)
		assert.Equal(t, "root/ORACLE-VPS-01/MONGODB", id)
		assert.Equal(t, []string{"ORACLE-VPS-01", "MONGODB"}, fc.created)
	})

	t.Run("given an empty path then returns the parent unchanged", func(t *testing.T) {
		fc := &fakeFolderCreator{}
		id, err := ResolveFolderPath(context.Background(), fc, "", "root")
		require.NoError(t, err)
		assert.Equal(t, "root", id)
		assert.Empty(t, fc.created)
	})

	t.Run("when a segment fails then returns the error and stops", func(t *testing.T) {
		fc := &fakeFolderCreator{failOn: "MONGODB"}
		id, err := ResolveFolderPath(context.Background(), fc, "ORACLE-VPS-01/MONGODB/DEEP", "root")
		require.Error(t, err)
		assert.Empty(t, id)
		assert.Equal(t, []string{"ORACLE-VPS-01"}, fc.created)
	})
}
