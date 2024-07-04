package gcp

import (
	"context"
	"fmt"
	"io"
	"log/slog"

	"cloud.google.com/go/storage"
	"google.golang.org/api/iterator"
)

type gcs struct {
	client *storage.Client
}

var (
	// Storage is the module level storage variable
	Storage *gcs
)

func init() {
	client, err := storage.NewClient(context.Background())
	if err != nil {
		slog.Warn("unable to create gcs client", "error", err)
		return
	}
	Storage = &gcs{client}
}

func (s *gcs) Close() {
	if s.client != nil {
		s.client.Close()
	}
}

func (s *gcs) Download(ctx context.Context, bucket, name string) (io.ReadCloser, string, error) {
	o := s.client.Bucket(bucket).Object(name)

	attrs, err := o.Attrs(ctx)
	if err != nil {
		return nil, "", err
	}

	r, err := o.NewReader(ctx)
	if err != nil {
		return nil, "", err
	}

	return r, attrs.ContentType, nil
}

func (s *gcs) Upload(ctx context.Context, r io.Reader, bucket, name, contentType string) error {
	b := s.client.Bucket(bucket)
	w := b.Object(name).NewWriter(ctx)
	defer w.Close()

	if contentType != "" {
		w.ContentType = contentType
	}

	if _, err := io.Copy(w, r); err != nil {
		return err
	}
	return nil
}

func (s *gcs) List(bucket string) error {
	it := s.client.Bucket(bucket).Objects(context.Background(), nil)
	for {
		attrs, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return err
		}
		fmt.Println(attrs.Name)
	}
	return nil
}
