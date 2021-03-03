package run

import (
	"context"
	"crypto/sha256"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/effxhq/vcs-connect/internal/effx"
	"github.com/effxhq/vcs-connect/internal/logger"
	"github.com/effxhq/vcs-connect/internal/model"

	"github.com/pkg/errors"

	"go.uber.org/zap"

	"github.com/effxhq/effx-cli/metadata"
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

func cloneMap(in map[string]string) map[string]string {
	out := make(map[string]string, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}

// Consumer is a stateless entity that ingests repositories from integrations.
type Consumer struct {
	EffxClient *effx.Client
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
func (c *Consumer) Consume(log *zap.Logger, repository *model.Repository) (err error) {
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

		// gets the dir where the effx file is at
		// for example /src/stuff/effx.yaml -> /src/stuff/
		effxDir := workDir + "/" + filepath.Dir(effxYAMLFile)

		result, err := metadata.InferMetadata(effxDir)
		if err != nil {
			log.Error("failed to infer langugage",
				zap.String("filPath", effxYAMLFile),
				zap.Error(err))
		}

		body, err := ioutil.ReadFile(path.Join(workDir, effxYAMLFile))
		if err != nil {
			if log != nil {
				log.Error("failed to read effx.yaml file",
					zap.String("filPath", effxYAMLFile),
					zap.Error(err))
			}
			continue
		}

		// set common annotations for this yaml
		tags := cloneMap(repository.Tags)
		annotations := cloneMap(repository.Annotations)
		inferredTags := []string{}

		if result != nil {
			if result.Language != "" {
				languageTag := "language"
				tags[languageTag] = result.Language
				inferredTags = append(inferredTags, languageTag)
			}

			if result.Language != "" && result.Version != "" {
				langVersionTag := strings.ToLower(result.Language)
				tags[langVersionTag] = strings.ToLower(result.Version)
				inferredTags = append(inferredTags, langVersionTag)

				fmt.Println("daniel", result.Language, " ", result.Version)
			}
		}

		annotations["effx.io/source"] = "vcs-connect"
		annotations["effx.io/repository"] = cloneURL
		annotations["effx.io/file-path"] = effxYAMLFile
		annotations["effx.io/inferred-tags"] = strings.Join(inferredTags, ",")

		err = c.EffxClient.Sync(&effx.SyncRequest{
			FileContents: string(body),
			Tags:         tags,
			Annotations:  annotations,
		})

		if err != nil {
			if log != nil {
				log.Error("failed to synx effx.yaml file",
					zap.String("filPath", effxYAMLFile),
					zap.Error(err))
			}
			continue
		}

		log.Info("successfully updated effx.yaml file",
			zap.String("filePath", effxYAMLFile))
	}

	err = c.EffxClient.DetectServices(workDir)
	if err != nil {
		log.Error("failed to detect services", zap.Error(err))
	}

	return nil
}

// Run consumes repositories from the data channel until the program is shutdown.
func (c *Consumer) Run(ctx context.Context, data chan *model.Repository) error {
	log := logger.MustGetFromContext(ctx)

	for {
		select {
		case <-ctx.Done():
			return nil
		case repository := <-data:
			if err := c.Consume(log, repository); err != nil {
				log.Error("failed to consume repository",
					zap.String("repository", repository.CloneURL),
					zap.Error(err))
			}
		}
	}
}
