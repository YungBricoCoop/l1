// SPDX-FileCopyrightText: 2026 Elwan Mayencourt <mayencourt@elwan.ch>
// SPDX-License-Identifier: MIT

package storage

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	appconfig "github.com/YungBricoCoop/l1/internal/config"
	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type PushInput struct {
	FilePath string
	Bucket   string
	Key      string
	Progress ProgressCallback
}

type PullInput struct {
	FilePath string
	Bucket   string
	Key      string
	Progress ProgressCallback
}

type ProgressUpdate struct {
	BytesTransferred int64
	TotalBytes       int64
	Done             bool
}

type ProgressCallback func(update ProgressUpdate)

type progressReader struct {
	reader           io.Reader
	totalBytes       int64
	bytesTransferred int64
	progress         ProgressCallback
}

func newProgressReader(reader io.Reader, totalBytes int64, progress ProgressCallback) *progressReader {
	if progress != nil {
		progress(ProgressUpdate{BytesTransferred: 0, TotalBytes: totalBytes, Done: false})
	}

	return &progressReader{
		reader:     reader,
		totalBytes: totalBytes,
		progress:   progress,
	}
}

func (r *progressReader) Read(p []byte) (int, error) {
	n, err := r.reader.Read(p)
	if n > 0 {
		r.bytesTransferred += int64(n)
		if r.progress != nil {
			r.progress(ProgressUpdate{
				BytesTransferred: r.bytesTransferred,
				TotalBytes:       r.totalBytes,
				Done:             false,
			})
		}
	}

	return n, err
}

func (r *progressReader) BytesTransferred() int64 {
	return r.bytesTransferred
}

func PushFile(ctx context.Context, cfg appconfig.Config, in PushInput) (string, error) {
	if err := cfg.ValidateForPush(in.Bucket); err != nil {
		return "", err
	}

	bucket := resolveBucket(cfg, in.Bucket)
	objectKey := resolveObjectKey(in.FilePath, in.Key)

	s3Client, err := buildS3Client(ctx, cfg)
	if err != nil {
		return "", err
	}

	file, err := os.Open(in.FilePath)
	if err != nil {
		return "", fmt.Errorf("open file %s: %w", in.FilePath, err)
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return "", fmt.Errorf("stat file %s: %w", in.FilePath, err)
	}

	reader := newProgressReader(file, fileInfo.Size(), in.Progress)

	output, err := s3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:        aws.String(bucket),
		Key:           aws.String(objectKey),
		Body:          reader,
		ContentLength: aws.Int64(fileInfo.Size()),
	})
	if err != nil {
		return "", fmt.Errorf("put object to s3://%s/%s: %w", bucket, objectKey, err)
	}

	if in.Progress != nil {
		in.Progress(ProgressUpdate{
			BytesTransferred: reader.BytesTransferred(),
			TotalBytes:       fileInfo.Size(),
			Done:             true,
		})
	}

	if output.ETag == nil {
		return "", nil
	}

	return *output.ETag, nil
}

func PullFile(ctx context.Context, cfg appconfig.Config, in PullInput) error {
	if err := cfg.ValidateForPush(in.Bucket); err != nil {
		return err
	}

	bucket := resolveBucket(cfg, in.Bucket)
	objectKey := resolveObjectKey(in.FilePath, in.Key)

	s3Client, err := buildS3Client(ctx, cfg)
	if err != nil {
		return err
	}

	output, err := s3Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(objectKey),
	})
	if err != nil {
		return fmt.Errorf("get object from s3://%s/%s: %w", bucket, objectKey, err)
	}
	defer output.Body.Close()

	mkdirErr := os.MkdirAll(filepath.Dir(in.FilePath), 0o750)
	if mkdirErr != nil {
		return fmt.Errorf("create destination directory: %w", mkdirErr)
	}

	destination, err := os.OpenFile(in.FilePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o600)
	if err != nil {
		return fmt.Errorf("open destination file %s: %w", in.FilePath, err)
	}
	defer destination.Close()

	totalBytes := int64(0)
	if output.ContentLength != nil {
		totalBytes = *output.ContentLength
	}

	reader := newProgressReader(output.Body, totalBytes, in.Progress)

	written, err := io.Copy(destination, reader)
	if err != nil {
		return fmt.Errorf("write destination file %s: %w", in.FilePath, err)
	}

	if in.Progress != nil {
		in.Progress(ProgressUpdate{
			BytesTransferred: written,
			TotalBytes:       totalBytes,
			Done:             true,
		})
	}

	return nil
}

func resolveBucket(cfg appconfig.Config, override string) string {
	bucket := strings.TrimSpace(override)
	if bucket != "" {
		return bucket
	}
	return strings.TrimSpace(cfg.S3.DefaultBucket)
}

func resolveObjectKey(filePath, override string) string {
	objectKey := strings.TrimSpace(override)
	if objectKey != "" {
		return objectKey
	}
	return filepath.Base(filePath)
}

func buildS3Client(ctx context.Context, cfg appconfig.Config) (*s3.Client, error) {
	accessKey, err := appconfig.ResolveSecretValue(cfg.S3.AccessKey)
	if err != nil {
		return nil, err
	}

	secretKey, err := appconfig.ResolveSecretValue(cfg.S3.SecretKey)
	if err != nil {
		return nil, err
	}

	region := strings.TrimSpace(cfg.S3.Region)
	if region == "" {
		region = "us-east-1"
	}

	loadOptions := []func(*awsconfig.LoadOptions) error{awsconfig.WithRegion(region)}
	if accessKey != "" || secretKey != "" {
		if accessKey == "" || secretKey == "" {
			return nil, errors.New("both s3.access_key and s3.secret_key are required when one is set")
		}

		provider := credentials.NewStaticCredentialsProvider(accessKey, secretKey, "")
		loadOptions = append(loadOptions, awsconfig.WithCredentialsProvider(provider))
	}

	endpointURL := strings.TrimSpace(cfg.S3.URL)

	awsCfg, err := awsconfig.LoadDefaultConfig(ctx, loadOptions...)
	if err != nil {
		return nil, fmt.Errorf("load aws config: %w", err)
	}

	s3Options := []func(*s3.Options){}
	if endpointURL != "" {
		s3Options = append(s3Options, func(o *s3.Options) {
			o.BaseEndpoint = aws.String(endpointURL)
		})
	}
	s3Options = append(s3Options, func(o *s3.Options) {
		if endpointURL != "" {
			o.UsePathStyle = true
		}
	})

	return s3.NewFromConfig(awsCfg, s3Options...), nil
}
