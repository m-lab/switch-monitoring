package config

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/url"
	"testing"

	"github.com/fsouza/fake-gcs-server/fakestorage"
	"github.com/m-lab/go/rtx"
)

func Test_gcsProvider_Get(t *testing.T) {
	// Start a GCS test server with a single bucket/file and no listener
	// so the test does not use the network stack.
	data, err := ioutil.ReadFile("testdata/abc01.conf")
	rtx.Must(err, "Cannot read test file")

	objects := []fakestorage.Object{
		{
			BucketName: "test",
			Content:    data,
			Name:       "abc01.conf",
		},
	}
	server, err := fakestorage.NewServerWithOptions(fakestorage.Options{
		InitialObjects: objects,
		NoListener:     true,
	})
	rtx.Must(err, "Cannot create fake GCS server")

	client := server.Client()
	provider := GCSProvider{
		bucket:   "test",
		client:   client,
		filename: "abc01.conf",
	}

	// Get an existing file.
	content, err := provider.Get(context.Background())
	if err != nil {
		t.Errorf("Get(): cannot get file: %v", err)
	}
	if bytes.Compare(content, data) != 0 {
		t.Errorf("Get(): content does not match")
	}

	// Trigger a reading error.
	readAll = func(io.Reader) ([]byte, error) {
		return nil, fmt.Errorf("error")
	}
	content, err = provider.Get(context.Background())
	if err == nil {
		t.Errorf("Get(): expected err, got nil")
	}
	if content != nil {
		t.Errorf("Get(): unexpected content returned")
	}

	// Get a non-existing file.
	provider.filename = "test.conf"
	content, err = provider.Get(context.Background())
	if err == nil {
		t.Errorf("Get(): expected err, got nil")
	}
	if content != nil {
		t.Errorf("Get(): unexpected content returned")
	}
}

func TestFromURL(t *testing.T) {
	// Pass an URL with an unsupported scheme.
	u, err := url.Parse("http://test")
	rtx.Must(err, "Cannot parse test URL")
	gcs, err := FromURL(context.Background(), u)
	if err != ErrUnsupportedURLScheme {
		t.Errorf("FromURL(): expected err, got nil or wrong error type")
	}
	if gcs != nil {
		t.Errorf("FromURL(): expected nil, got provider")
	}

	// Pass an URL with a bucket but no filename.
	u, err = url.Parse("gs://bucket/")
	rtx.Must(err, "Cannot parse test URL")
	gcs, err = FromURL(context.Background(), u)
	if err != ErrNoFilenameInURL {
		t.Errorf("FromURL(): expected err, got nil or wrong error type")
	}
	if gcs != nil {
		t.Errorf("FromURL(): expected nil, got provider")
	}

	// Pass a valid URL.
	u, err = url.Parse("gs://bucket/abc01.conf")
	rtx.Must(err, "Cannot parse test URL")
	gcs, err = FromURL(context.Background(), u)
	if err != nil {
		t.Errorf("FromURL(): unexpected error: %v", err)
	}
	if gcs.bucket != "bucket" || gcs.filename != "abc01.conf" {
		t.Errorf("FromURL() did not return the expected gcsProvider")
	}
}
