package storage

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"time"
)

func GeneratePresignedURL(client *s3.Client, key string) (string, error) {

	presignClient := s3.NewPresignClient(client)

	req, err := presignClient.PresignGetObject(
		context.TODO(),
		&s3.GetObjectInput{
			Bucket: aws.String("candycane"),
			Key:    aws.String(key),
		},
		s3.WithPresignExpires(4*time.Hour),
	)

	if err != nil {
		return "", err
	}

	return req.URL, nil
}

func NewClient() (*s3.Client, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO(),config.WithRegion("us-east-1"),)
	if err != nil {
		return nil, err
	}

	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String("https://sgp1.digitaloceanspaces.com") // your DO Spaces endpoint
	})

	return client, nil
}
