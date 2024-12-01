package builder

import (
	"bytes"
	"context"
	"path/filepath"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

type AwsManager struct {
	Bucket string
}

// IsReady checks to see if the AWS setup is ready to be used.
//
// It does this by getting an S3 client and then trying to get the CORS
// configuration of the bucket. If the client can't be created or the CORS
// configuration can't be retrieved, it returns false. Otherwise, it returns
// true.
func (a AwsManager) IsReady() bool {
	log.Info().Msg("Making sure AWS is ready.")

	client, err := a.GetClient()
	if err != nil {
		log.Error().Err(err).Msg("Error getting client")
		return false
	}
	_, err = client.GetBucketCors(context.TODO(), &s3.GetBucketCorsInput{Bucket: &a.Bucket})
	if err != nil {
		log.Error().Err(err).Msg("Error getting bucket cors")
		return false
	}

	return true
}

// GetClient returns a new S3 client using the default AWS configuration with
// the region set to us-east-1. If the default configuration can't be loaded,
// it returns an error.
func (a AwsManager) GetClient() (*s3.Client, error) {
	cfg, err := awsConfig.LoadDefaultConfig(context.TODO(), awsConfig.WithRegion("us-east-1"))
	if err != nil {
		log.Error().Err(err).Msg("Error loading AWS config")
		return nil, err
	}

	return s3.NewFromConfig(cfg), nil
}

// UploadFile uploads the given file to the given bucket with the given key.
// It uploads the file with public-read permissions. If there is an error
// uploading the file, it logs an error and returns the error.
func (a AwsManager) UploadFile(directory string, fileName string, file []byte) error {
	log.Info().Str("fileName", fileName).Msg("Uploading file to S3.")

	ctx := context.Background()
	objectKey := aws.String(fileName)

	client, err := a.GetClient()
	if err != nil {
		log.Error().Err(err).Msg("Error getting client")
		return err
	}

	if directory != "" {
		objectKey = aws.String(filepath.Join(directory, fileName))
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

// DeleteFile deletes the given file from the given bucket. It logs an error
// and returns the error if there is an error deleting the file.
func (a AwsManager) DeleteFile(fileName string) error {
	log.Info().Str("fileName", fileName).Msg("Deleting file from S3.")

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
