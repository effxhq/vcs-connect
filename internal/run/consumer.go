package run

import (
	"context"
	"crypto/sha256"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/effxhq/vcs-connect/internal/logger"
	"github.com/effxhq/vcs-connect/internal/model"

	"github.com/pkg/errors"

	"go.uber.org/zap"

	"gopkg.in/src-d/go-billy.v4/osfs"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing/cache"
	"gopkg.in/src-d/go-git.v4/plumbing/transport"
	"gopkg.in/src-d/go-git.v4/storage/filesystem"
)

var (
	// matches *.effx.yaml, effx.yaml, *.effx.yml, effx.yml
	effxYAMLPattern, _ = regexp.Compile("^(.+\\.)?effx\\.ya?ml$")
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

			if effxYAMLPattern.MatchString(fileName) {
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
	defer os.RemoveAll(workDir)

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

func (c *Consumer) Run(ctx context.Context, data chan *model.Repository) error {
	log := logger.MustGetFromContext(ctx)

	for {
		select {
		case <-ctx.Done():
			return nil
		case repository := <-data:
			if err := c.Consume(repository); err != nil {
				log.Error("failed to consume repository",
					zap.String("repository", repository.CloneURL),
					zap.Error(err))
			}
		}
	}
}
