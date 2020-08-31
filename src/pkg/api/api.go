package api

import (
	"net/url"
	"time"
	"context"
)

type Metadata struct {
	Labels map[string]string
	Annotations map[string]string
	Timestamp time.Time
}

type Index interface {
	Write(ctx context.Context, location url.URL, metadata Metadata) error
	Read(ctx context.Context, location url.URL) (Metadata, error)
//	Search(ctx context.Context, selector Label) ([]url.URL, error)
}

// Externalizer turns a local file:// URL into an external http:// URL.
type Externalizer interface {
	Externalize(loc url.URL) (url.URL, error)
}

