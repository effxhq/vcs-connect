package run_test

import (
	"os"
	"path"
	"testing"

	"github.com/effxhq/vcs-connect/internal/run"

	"github.com/stretchr/testify/require"
)

func TestConsumer_SetupFS(t *testing.T) {
	tmp := t.TempDir()
	defer os.Remove(tmp)

	c := &run.Consumer{}

	// TODO: Use self instead of a separate random repo?
	err := c.SetupFS(tmp, "https://github.com/effxhq/effx-sync-action.git")
	require.NoError(t, err)

	_, err = os.Stat(path.Join(tmp, "LICENSE"))
	require.NoError(t, err)

	_, err = os.Stat(path.Join(tmp, "README.md"))
	require.NoError(t, err)
}

func TestConsumer_FindEffxYAML(t *testing.T) {
	c := &run.Consumer{}

	workDir := path.Join("..", "..", "hack", "run")

	files, err := c.FindEffxYAML(workDir)
	require.NoError(t, err)

	require.Len(t, files, 3)
	require.Contains(t, files, "effx.yaml")
	require.Contains(t, files, "effx.yml")
	require.Contains(t, files, "prefixed.effx.yaml")
}

func TestConsumer_InferLangugage(t *testing.T) {
	c := &run.Consumer{}

	workDir := path.Join("./")

	lang, err := c.InferLanguage(workDir)
	require.NoError(t, err)

	require.Contains(t, lang, "go")
}
