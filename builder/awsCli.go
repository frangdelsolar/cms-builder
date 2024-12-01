package builder

import (
	"bytes"
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

type AwsManager struct {
	Bucket string
}

func (a AwsManager) IsReady() bool {
	log.Info().Msg("Checking if AWS is ready")

	client, err := a.GetClient()
	if err != nil {
		log.Error().Err(err).Msg("Error getting client")
		return false
	}
	output, err := client.GetBucketCors(context.TODO(), &s3.GetBucketCorsInput{Bucket: &a.Bucket})
	if err != nil {
		log.Error().Err(err).Msg("Error getting bucket cors")
		return false
	}

	log.Debug().Interface("output", output).Msg("Bucket cors is ready")
	return true
}

func (a AwsManager) GetClient() (*s3.Client, error) {
	cfg, err := awsConfig.LoadDefaultConfig(context.TODO(), awsConfig.WithRegion("us-east-1"))
	if err != nil {
		log.Error().Err(err).Msg("Error loading AWS config")
		return nil, err
	}

	return s3.NewFromConfig(cfg), nil
}

func (a AwsManager) UploadFile(fileName string, file []byte) error {
	log.Info().Msg("Uploading file to S3.")

	ctx := context.Background()
	objectKey := aws.String(fileName)

	client, err := a.GetClient()
	if err != nil {
		log.Error().Err(err).Msg("Error getting client")
		return err
	}
	_, err = client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(a.Bucket),
		Key:    objectKey,
		Body:   bytes.NewReader(file),
		ACL:    types.ObjectCannedACL("public-read"),
	})
	if err != nil {
		log.Error().Err(err).Msg("Error uploading file to S3")
		return err
	}

	return nil
}

func (a AwsManager) DeleteFile(fileName string) error {
	log.Info().Msg("Deleting file from S3.")

	ctx := context.Background()
	objectKey := aws.String(fileName)

	client, err := a.GetClient()
	if err != nil {
		log.Error().Err(err).Msg("Error getting client")
		return err
	}
	_, err = client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(a.Bucket),
		Key:    objectKey,
	})
	if err != nil {
		log.Error().Err(err).Msg("Error deleting file from S3")
		return err
	}

	return nil
}
