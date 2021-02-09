package run

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/go-enry/go-enry/v2"
	"github.com/pkg/errors"
)

// InferLanguage detects the programming used in the provided work directory .
func (c *Consumer) InferLanguage(workDir string) ([]string, error) {
	langs := make([]string, 0)

	collector := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		} else if !info.IsDir() {
			fileName := filepath.Base(path)

			content, err := ioutil.ReadFile(fileName)
			if err != nil {
				return err
			}

			// infers langugage from extension, and code content.
			lang := enry.GetLanguage(fileName, content)
			if lang != "" {
				langs = append(langs, lang)
			}
		}
		return nil
	}

	err := filepath.Walk(workDir, collector)
	if err != nil {
		return nil, errors.Wrap(err, "encountered error inferring langugages")
	}

	return langs, err
}
