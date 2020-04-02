package config

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net/url"
	"time"

	"cloud.google.com/go/storage"
	"google.golang.org/api/iterator"
)

const gcsTimeout = time.Second * 15

var (
	// ErrUnsupportedURLScheme is the error returned when the URL scheme
	// is unsupported.
	ErrUnsupportedURLScheme = errors.New("Unsupported URL scheme")

	readAll = ioutil.ReadAll
)

// GCSProvider is a Google Cloud Storage provider.
type GCSProvider struct {
	bucket string
	client *storage.Client
}

// get returns the content of the object located at filename from the
// configured GCS bucket.
func (g GCSProvider) get(ctx context.Context, filename string) ([]byte,
	error) {

	ctx, cancel := context.WithTimeout(ctx, gcsTimeout)
	defer cancel()
	obj := g.client.Bucket(g.bucket).Object(filename)
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

// GetLatestConfig downloads the latest configuration file for a given site
// from GCS
func (g GCSProvider) GetLatestConfig(ctx context.Context) ([]byte, error) {
	it := g.client.Bucket(g.bucket).Objects(ctx, nil)
	for {
		attrs, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		fmt.Println(attrs.Name)
	}
	return nil, nil
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
	return &GCSProvider{
		client: client,
		bucket: u.Host,
	}, err
}
