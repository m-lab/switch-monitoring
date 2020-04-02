package config

import (
	"context"
	"errors"
	"io/ioutil"
	"net/url"
	"strings"
	"time"

	"cloud.google.com/go/storage"
)

const gcsTimeout = time.Second * 15

var (
	// ErrUnsupportedURLScheme is the error returned when the URL scheme
	// is unsupported.
	ErrUnsupportedURLScheme = errors.New("Unsupported URL scheme")

	// ErrNoFilenameInURL is the error returned when the URL does not contain
	// any recognizable filename.
	ErrNoFilenameInURL = errors.New("Bad URL, no filename detected")
	readAll            = ioutil.ReadAll
)

// GCSProvider is a Google Cloud Storage provider.
type GCSProvider struct {
	bucket, filename string
	client           *storage.Client
}

// Get returns the content of the object located at filename from the
// configured GCS bucket.
func (g GCSProvider) Get(ctx context.Context) ([]byte, error) {
	ctx, cancel := context.WithTimeout(ctx, gcsTimeout)
	defer cancel()
	obj := g.client.Bucket(g.bucket).Object(g.filename)
	r, err := obj.NewReader(ctx)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	data, err := readAll(r)
	if err != nil {
		return nil, err
	}

	return data, nil
}

// FromURL returns a new GCSProvider based on the passed-in URL. The only
// supported URL scheme is gs://bucket/filename. Whether the path contained
// in the URL is valid isn't known until the Get() method of the returned
// GCSProvider is called. Unsupported URL schemes cause this to return
// ErrUnsupportedURLScheme.
func FromURL(ctx context.Context, u *url.URL) (*GCSProvider, error) {
	if u.Scheme != "gs" {
		return nil, ErrUnsupportedURLScheme
	}

	client, err := storage.NewClient(ctx)
	filename := strings.TrimPrefix(u.Path, "/")
	if len(filename) == 0 {
		return nil, ErrNoFilenameInURL
	}
	return &GCSProvider{
		client:   client,
		bucket:   u.Host,
		filename: filename,
	}, err
}
