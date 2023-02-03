/*
 * File: datastore.go
 * Project: clients
 * File Created: Saturday, 28th January 2023 10:46:32 am
 * Author: Mark Mester (mmester6016@gmail.com)
 * -----
 * Last Modified: Saturday, 28th January 2023 7:35:45 pm
 * Modified By: Mark Mester (mmester6016@gmail.com>)
 */
package clients

import (
	"context"
	"os"
	"path"
	"strings"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/pkg/errors"
)

type DataStore interface {
	Get(key string) (data []byte, err error)
	Set(key string, data []byte) error
}

// ============= S3 Client ============= //

type S3Client struct {
	*s3.Client

	bucket string
}

func NewS3Client(bucket string) (*S3Client, error) {
	sdkConfig, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return nil, errors.Wrapf(err, "couldn't load default configuration")
	}

	return &S3Client{
		Client: s3.NewFromConfig(sdkConfig),
		bucket: bucket,
	}, nil

}

func (s *S3Client) Get(key string) (data []byte, err error) {
	return nil, nil
}

func (s *S3Client) Set(key string, data []byte) error {
	return nil
}

// ============= Local Client ============= //

type LocalClient struct {
	path string
}

func NewLocalClient(path string) (*LocalClient, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, errors.Wrapf(err, "path does not exists")
	}

	return &LocalClient{
		path: path,
	}, nil
}

func (l *LocalClient) Get(key string) (data []byte, err error) {
	fileBytes, err := os.ReadFile(path.Join(l.path, key))
	if err != nil {
		return nil, err
	}
	return fileBytes, nil
}

func (l *LocalClient) Set(key string, data []byte) error {
	return os.WriteFile(key, data, 0744)
}

// ============= NooP Client ============= //

type NooPClient struct{}

func (n *NooPClient) Get(_ string) (data []byte, err error) {
	return nil, os.ErrNotExist
}

func (n *NooPClient) Set(_ string, _ []byte) error {
	return nil
}

func NewDatastore(path string) (DataStore, error) {

	// S3 path?
	if strings.HasPrefix(path, "s3://") {
		return NewS3Client(strings.TrimPrefix(path, "s3://"))
	}

	// Local path?
	if info, err := os.Stat(path); err == nil && info.IsDir() {
		return NewLocalClient(path)
	}

	// Default to in memory
	return &NooPClient{}, nil
}
