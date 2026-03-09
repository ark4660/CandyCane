package storage

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"io"
	"bytes"
	"log"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

func UploadData(dataStream io.ReadCloser, key string, s3Client *s3.Client) error {
	_, err := s3Client.PutObject(context.TODO(), &s3.PutObjectInput{
    		Bucket: aws.String("candycane"),
      		Key:    aws.String("videos/" + key + "/video"),
       		Body:   dataStream,
         	ContentType: aws.String("video/mp4"),
         	ACL:    "private",
	})
	if err != nil {
		panic(err)
	}
	return nil
}

func UploadThumbnail(dataStream io.ReadCloser, key string, s3Client *s3.Client) error {
	_, err := s3Client.PutObject(context.TODO(), &s3.PutObjectInput{
    		Bucket: aws.String("candycane"),
      		Key:    aws.String("videos/" + key + "/thumbnail"),
       		Body:   dataStream,
         	ContentType: aws.String("image/jpg"),
         	ACL:    "public-read",
	})
	if err != nil {
		panic(err)
	}
	return nil
}

func UploadTranscriptJson(dataStream io.ReadCloser, key string, s3Client *s3.Client) error {
	buf, _ := io.ReadAll(dataStream)
	size := int64(len(buf))
	_, err := s3Client.PutObject(context.TODO(), &s3.PutObjectInput{
    		Bucket: aws.String("candycane"),
      		Key:    aws.String("videos/" + key + "/transcript"),
       		Body:   bytes.NewReader(buf),
         	ContentLength: &size,
          	ACL:    "public-read",
	})
	log.Println(err)
	return err
}

func UploadTranscriptVtt(dataStream io.ReadCloser, key string, s3Client *s3.Client) error {
	_, err := s3Client.PutObject(context.TODO(), &s3.PutObjectInput{
    		Bucket: aws.String("candycane"),
      		Key:    aws.String("videos/" + key + "/transcriptVtt"),
       		Body:   dataStream,
         	ContentType: aws.String("text/plain"),
         	ACL:    types.ObjectCannedACLPublicRead,
	})
	log.Println(err)
	return err
}
