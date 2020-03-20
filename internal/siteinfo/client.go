package siteinfo

import (
	"fmt"
	"io/ioutil"

	"github.com/m-lab/switch-monitoring/internal"
)

const (
	baseURLFormat = "https://siteinfo.%s.measurementlab.net/v1/"
)

// Client is a Siteinfo client.
type Client struct {
	ProjectID  string
	httpClient internal.HTTPProvider
}

// New returns a new Siteinfo client wrapping the provided *http.Client.
func New(projectID string, httpClient internal.HTTPProvider) *Client {
	return &Client{
		ProjectID:  projectID,
		httpClient: httpClient,
	}
}

// Switches fetches the switches.json output format and returns its content.
func (s Client) Switches() ([]byte, error) {
	url := fmt.Sprintf(baseURLFormat+"sites/switches.json", s.ProjectID)

	resp, err := s.httpClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}
