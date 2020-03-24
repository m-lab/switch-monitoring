package siteinfo

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"testing"
)

const switchesPath = "testdata/switches.json"

// fileReaderProvider implements a HTTPProvider but the response's content
// comes from a configurable file.
type fileReaderProvider struct {
	path           string
	mustFail       bool
	mustFailToRead bool
}

func (prov fileReaderProvider) Get(string) (*http.Response, error) {
	if prov.mustFail {
		return nil, fmt.Errorf("error")
	}

	// Note: it's the caller's responsibility to call Body.Close().
	f, _ := os.Open(prov.path)

	var body io.ReadCloser
	if prov.mustFailToRead {
		defer f.Close()
		body = &mockReadCloser{}
	} else {
		body = ioutil.NopCloser(bufio.NewReader(f))
	}
	return &http.Response{
		Body:       body,
		StatusCode: http.StatusOK,
	}, nil
}

// mockReadCloser is ReadCloser that fails.
type mockReadCloser struct{}

func (mockReadCloser) Read(p []byte) (n int, err error) {
	return 0, fmt.Errorf("error")
}

func (mockReadCloser) Close() error {
	return nil
}

//
// Tests start here.
//

func TestNew(t *testing.T) {
	client := New("project", http.DefaultClient)
	if client == nil {
		t.Errorf("New() returned nil.")
	}
}

func TestClient_Switches(t *testing.T) {
	prov := &fileReaderProvider{
		path: "testdata/switches.json",
	}
	client := New("test", prov)

	testData, err := ioutil.ReadFile(switchesPath)
	if err != nil {
		t.Errorf("Cannot read test data from %v", switchesPath)
	}

	// This should return the content of the test file.
	res, err := client.Switches()
	if err != nil {
		t.Errorf("Switches() returned err: %v", err)
	}

	if bytes.Compare(res, testData) != 0 {
		t.Errorf("Switches(): expected: %v, got %v",
			testData, res)
	}

	// Make the HTTP client fail.
	prov.mustFail = true
	res, err = client.Switches()
	if err == nil {
		t.Errorf("Switches(): expected err, got nil.")
	}
	prov.mustFail = false

	// Make reading the response body fail.
	prov.mustFailToRead = true
	res, err = client.Switches()
	if err == nil {
		t.Errorf("Switches(): expected err, got nil.")
	}
	prov.mustFailToRead = false
}
