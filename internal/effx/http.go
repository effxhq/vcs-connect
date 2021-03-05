package effx

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/effxhq/effx-cli/discover"
	"github.com/thoas/go-funk"
)

// New returns an effx Client encapsulating operations with the API
func New(cfg *Configuration) (*Client, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return &Client{cfg}, nil
}

// SyncError contains information provided when an error occurs
type SyncError struct {
	Message string `json:"message,omitempty"`
}

// SyncRequest contains a config blob for indexing.
type SyncRequest struct {
	FileContents string            `json:"fileContents,omitempty"`
	Tags         map[string]string `json:"tags,omitempty"`
	Annotations  map[string]string `json:"annotations,omitempty"`
}

// Client encapsulates communication with the API.
type Client struct {
	cfg *Configuration
}

// Returns if a given feature is disabled.
// example: LANGUAGE_DETECTION
func (c *Client) IsFeatureDisabled(feature string) bool {
	list := strings.Split(c.cfg.Disable, ",")
	return funk.ContainsString(list, feature)
}

// Sync attempts to synchronize provided contents with the upstream api.
func (c *Client) Sync(syncRequest *SyncRequest) error {
	body, err := json.Marshal(syncRequest)
	if err != nil {
		return err
	}

	reader := bytes.NewReader(body)

	endpoint := url.URL{
		Scheme: "https",
		Host:   c.cfg.APIHost,
		Path:   "/v2/config",
	}

	req, err := http.NewRequest("PUT", endpoint.String(), reader)
	if err != nil {
		return err
	}
	req.Header.Add("content-type", "application/json")
	req.Header.Add("x-effx-api-key", c.cfg.APIKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		syncErr := &SyncError{}
		err = json.NewDecoder(resp.Body).Decode(syncErr)
		if err != nil {
			return err
		}

		return fmt.Errorf(syncErr.Message)
	}

	return nil
}

// DetectServices attempts to detect services based on repo work dir.
func (c *Client) DetectServices(workDir string) error {
	return discover.DetectServicesFromWorkDir(workDir, c.cfg.APIKey, "vcs-connect")
}
