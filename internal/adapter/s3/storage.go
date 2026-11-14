// Package s3 提供 S3 兼容对象存储适配器，默认关闭。
package s3

import (
	"bytes"
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	awss3 "github.com/aws/aws-sdk-go-v2/service/s3"

	"facial-liveness-verification-go/internal/config"
	"facial-liveness-verification-go/internal/domain"
)

// Storage 是 S3 兼容对象存储。
type Storage struct {
	cfg     config.Storage
	client  *awss3.Client
	presign *awss3.PresignClient
}

// New 构造对象存储；未启用时不创建客户端，调用方会在使用时收到 503。
func New(ctx context.Context, cfg config.Storage) (*Storage, error) {
	s := &Storage{cfg: cfg}
	if !cfg.Enabled {
		return s, nil
	}
	awsCfg, err := awsconfig.LoadDefaultConfig(ctx,
		awsconfig.WithRegion(cfg.Region),
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(cfg.AccessKey, cfg.SecretKey, "")),
	)
	if err != nil {
		return nil, err
	}
	s.client = awss3.NewFromConfig(awsCfg, func(o *awss3.Options) {
		o.BaseEndpoint = aws.String(cfg.Endpoint)
		o.UsePathStyle = cfg.ForcePathStyle
	})
	s.presign = awss3.NewPresignClient(s.client)
	return s, nil
}

// Put 上传对象。
func (s *Storage) Put(ctx context.Context, key string, content []byte, contentType string) error {
	if err := s.ensure(); err != nil {
		return err
	}
	_, err := s.client.PutObject(ctx, &awss3.PutObjectInput{
		Bucket:      aws.String(s.cfg.Bucket),
		Key:         aws.String(key),
		Body:        bytes.NewReader(content),
		ContentType: aws.String(contentType),
	})
	return err
}

// Delete 删除对象。
func (s *Storage) Delete(ctx context.Context, key string) error {
	if err := s.ensure(); err != nil {
		return err
	}
	_, err := s.client.DeleteObject(ctx, &awss3.DeleteObjectInput{
		Bucket: aws.String(s.cfg.Bucket),
		Key:    aws.String(key),
	})
	return err
}

// PresignGet 生成短时效私有读 URL。
func (s *Storage) PresignGet(ctx context.Context, key string, ttl time.Duration) (string, error) {
	if err := s.ensure(); err != nil {
		return "", err
	}
	out, err := s.presign.PresignGetObject(ctx, &awss3.GetObjectInput{
		Bucket: aws.String(s.cfg.Bucket),
		Key:    aws.String(key),
	}, awss3.WithPresignExpires(ttl))
	if err != nil {
		return "", err
	}
	return out.URL, nil
}

func (s *Storage) ensure() error {
	if !s.cfg.Enabled {
		return domain.New(domain.CodeServiceUnavailable, "对象存储未启用（Storage:Enabled=false）")
	}
	return nil
}
