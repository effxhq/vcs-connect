package run

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/effxhq/vcs-connect/internal/model"

	"github.com/pkg/errors"

	"gopkg.in/src-d/go-billy.v4/osfs"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing/cache"
	"gopkg.in/src-d/go-git.v4/plumbing/transport"
	"gopkg.in/src-d/go-git.v4/storage/filesystem"
)

func s256(in string) string {
	return fmt.Sprintf("%x", sha256.Sum256([]byte(in)))
}

// Consumer is a stateless entity that ingests repositories from integrations.
type Consumer struct {
	ScratchDir string
	AuthMethod transport.AuthMethod
}

// SetupFS initializes the workspace with the corresponding git repository.
func (c *Consumer) SetupFS(workDir, cloneURL string) error {
	fs := osfs.New(workDir)
	gitfs, err := fs.Chroot(git.GitDirName)
	if err != nil {
		return errors.Wrapf(err, "failed to setup .git directory")
	}

	storage := filesystem.NewStorage(gitfs, cache.NewObjectLRUDefault())
	options := &git.CloneOptions{
		URL:   cloneURL,
		Depth: 1,
		Auth:  c.AuthMethod,
	}

	_, err = git.Clone(storage, fs, options)
	if err != nil {
		return errors.Wrapf(err, "failed to clone repository")
	}

	return nil
}

// FindEffxYAML searches the provided work directory for effx.yaml files.
func (c *Consumer) FindEffxYAML(workDir string) ([]string, error) {
	files := make([]string, 0)
	collector := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		} else if !info.IsDir() {
			fileName := filepath.Base(path)

			// matches *.effx.yaml and effx.yaml
			if fileName == "effx.yaml" || strings.HasSuffix(path, ".effx.yaml") {
				files = append(files, strings.TrimPrefix(path, workDir)[1:])
			}
		}
		return nil
	}

	err := filepath.Walk(workDir, collector)
	if err != nil {
		return nil, errors.Wrap(err, "encountered error locating effx.yaml files")
	}
	return files, err
}

// Consume attempts to index a repository for effx.yaml files
func (c *Consumer) Consume(repository *model.Repository) (err error) {
	cloneURL := repository.CloneURL
	workDir := path.Join(c.ScratchDir, s256(cloneURL))

	// clean up workspace
	defer os.Remove(workDir)

	err = c.SetupFS(workDir, cloneURL)
	if err != nil {
		return err
	}

	effxYAML, err := c.FindEffxYAML(workDir)
	if err != nil {
		return err
	}

	// parse and send to our API
	for _, effxYAMLFile := range effxYAML {
		fmt.Println(effxYAMLFile)
	}

	return nil
}
