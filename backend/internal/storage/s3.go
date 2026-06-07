// Package storage wraps an S3-compatible object store (MinIO in dev, Cloudflare
// R2 in prod). Browsers upload directly via pre-signed PUT URLs; the API only
// stores keys and resolves them to public URLs.
package storage

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type Config struct {
	Endpoint         string
	ExternalEndpoint string
	Region           string
	AccessKeyID      string
	SecretAccessKey  string
	Bucket           string
	PublicBaseURL    string
	UsePathStyle     bool
}

type Client struct {
	s3      *s3.Client
	presign *s3.PresignClient
	cfg     Config
}

func New(ctx context.Context, cfg Config) (*Client, error) {
	awsCfg, err := awsconfig.LoadDefaultConfig(ctx,
		awsconfig.WithRegion(cfg.Region),
		awsconfig.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(cfg.AccessKeyID, cfg.SecretAccessKey, ""),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("load aws config: %w", err)
	}

	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		if cfg.Endpoint != "" {
			o.BaseEndpoint = aws.String(cfg.Endpoint)
		}
		o.UsePathStyle = cfg.UsePathStyle
	})

	return &Client{
		s3:      client,
		presign: s3.NewPresignClient(client),
		cfg:     cfg,
	}, nil
}

// PresignPut returns a URL the browser can PUT to directly.
func (c *Client) PresignPut(ctx context.Context, key, contentType string, expires time.Duration) (string, error) {
	req, err := c.presign.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(c.cfg.Bucket),
		Key:         aws.String(key),
		ContentType: aws.String(contentType),
	}, s3.WithPresignExpires(expires))
	if err != nil {
		return "", fmt.Errorf("presign put: %w", err)
	}
	return c.rewriteExternal(req.URL), nil
}

// PublicURL resolves a stored key to a browser-facing GET URL.
func (c *Client) PublicURL(key string) string {
	if key == "" {
		return ""
	}
	if c.cfg.PublicBaseURL != "" {
		return strings.TrimRight(c.cfg.PublicBaseURL, "/") + "/" + key
	}
	return key
}

func (c *Client) Delete(ctx context.Context, key string) error {
	_, err := c.s3.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(c.cfg.Bucket),
		Key:    aws.String(key),
	})
	return err
}

// rewriteExternal swaps the internal endpoint host (e.g. minio:9000) for the
// browser-reachable one (e.g. localhost:9000) while keeping the signature valid.
func (c *Client) rewriteExternal(url string) string {
	if c.cfg.ExternalEndpoint == "" || c.cfg.Endpoint == "" {
		return url
	}
	return strings.Replace(url, c.cfg.Endpoint, c.cfg.ExternalEndpoint, 1)
}
