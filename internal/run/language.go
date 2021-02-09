package run

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-enry/go-enry/v2"
	"github.com/pkg/errors"
)

func determineMostCommonLangugage(languageCount map[string]int) string {
	max := 0
	mostCommonLang := ""

	for key, value := range languageCount {
		if max < value {
			max = value
			mostCommonLang = key
		}
	}

	return mostCommonLang
}

// InferLanguage detects the programming used in the provided work directory .
func (c *Consumer) InferLanguage(workDir string) (string, error) {
	languageCount := map[string]int{}

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
			lang := strings.ToLower(enry.GetLanguage(fileName, content))

			if count, ok := languageCount[lang]; ok {
				languageCount[lang] = count + 1
			} else {
				languageCount[lang] = 1
			}
		}
		return nil
	}

	err := filepath.Walk(workDir, collector)
	if err != nil {
		return "", errors.Wrap(err, "encountered error inferring langugages")
	}

	return determineMostCommonLangugage(languageCount), err
}
