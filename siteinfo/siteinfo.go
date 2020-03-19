package siteinfo

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

const (
	baseURLFormat    = "https://siteinfo.%s.measurementlab.net/v1/"
	switchHostFormat = "s1.%s.measurement-lab.org"
)

type Siteinfo struct {
	ProjectID string
}

func (s Siteinfo) Switches() ([]string, error) {
	sites, err := s.Sites()
	if err != nil {
		return nil, err
	}

	for i, s := range sites {
		sites[i] = fmt.Sprintf(switchHostFormat, s)
	}

	return sites, nil
}

func (s Siteinfo) Sites() ([]string, error) {
	var switches map[string]interface{}
	switchesJSON, err := s.getSwitches()
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(switchesJSON, &switches)
	if err != nil {
		return nil, err
	}

	if len(switches) == 0 {
		return nil, fmt.Errorf("the retrieved switches list is empty")
	}

	keys := make([]string, len(switches))

	i := 0
	for k := range switches {
		keys[i] = k
		i++
	}

	return keys, nil
}

func (s Siteinfo) getSwitches() ([]byte, error) {
	url := fmt.Sprintf(baseURLFormat+"sites/switches.json", s.ProjectID)

	resp, err := http.Get(url)
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
