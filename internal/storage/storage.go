package storage

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type Storage struct {
	client *s3.Client
	bucket string
}

func NewStorage(accountID, accessKeyID, accessKeySecret, bucketName string) *Storage {
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKeyID, accessKeySecret, "")),
		config.WithRegion("auto"),
	)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(fmt.Sprintf("https://%s.r2.cloudflarestorage.com", accountID))
		o.UsePathStyle = true
	})

	return &Storage{
		client: client,
		bucket: bucketName,
	}
}

func (s *Storage) ListObjects(ctx context.Context) (*s3.ListObjectsV2Output, error) {
	page, err := s.client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
		Bucket: aws.String(s.bucket),
	})
	if err != nil {
		return nil, fmt.Errorf("Error listing objects: %v", err)
	}
	return page, nil
}

func (s *Storage) ListFileLinks(ctx context.Context) ([]FileDataResponse, error) {
	page, _ := s.ListObjects(ctx)

	var data []FileDataResponse
	for _, obj := range page.Contents {
		url, _ := s.GetFileLink(ctx, *obj.Key)
		data = append(data, FileDataResponse{Key: *obj.Key, URL: url})
	}

	return data, nil
}

func (s *Storage) GetFileLink(ctx context.Context, objKey string) (string, error) {
	presignedClient := s3.NewPresignClient(s.client)
	presignUrl, err := presignedClient.PresignGetObject(ctx,
		&s3.GetObjectInput{
			Bucket: aws.String(s.bucket),
			Key:    aws.String(objKey),
		},
		s3.WithPresignExpires(5*time.Minute),
	)
	if err != nil {
		return "", fmt.Errorf("Error getting file from storage: %w", err)
	}

	return presignUrl.URL, nil
}

func (s *Storage) GetUploadURL(ctx context.Context, data FileDataRequest) (string, error) {
	presignedClient := s3.NewPresignClient(s.client)
	presignedReq, err := presignedClient.PresignPutObject(ctx,
		&s3.PutObjectInput{
			Bucket: &s.bucket,
			Key:    &data.Key,
		},
		s3.WithPresignExpires(5*time.Minute),
	)
	if err != nil {
		log.Println(fmt.Errorf("Error getting upload URL: %w", err))
		return "", fmt.Errorf("Error getting upload URL: %w", err)
	}

	return presignedReq.URL, nil
}

func (s *Storage) DeleteFile(ctx context.Context, key string) error {
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: &s.bucket,
		Key:    aws.String(key),
	})

	log.Printf("tried to delete: %s", key)
	return err
}
