package clients

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"path/filepath"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"

	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/logger"
	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/models"
	. "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/models"
)

type AwsManager struct {
	Bucket    string
	Region    string
	AccessKey string
	SecretKey string
}

// IsReady checks to see if the AWS setup is ready to be used.
//
// It does this by getting an S3 client and then trying to get the CORS
// configuration of the bucket. If the client can't be created or the CORS
// configuration can't be retrieved, it returns false. Otherwise, it returns
// true.
func (a *AwsManager) IsReady() bool {
	client, err := a.GetClient()
	if err != nil {
		return false
	}
	_, err = client.GetBucketCors(context.TODO(), &s3.GetBucketCorsInput{Bucket: &a.Bucket})
	if err != nil {
		return false
	}

	return true
}

// GetClient returns a new S3 client using the default AWS configuration with
// the region set to us-east-1. If the default configuration can't be loaded,
// it returns an error.
func (a *AwsManager) GetClient() (*s3.Client, error) {
	cfg := aws.Config{
		Region: a.Region,
		Credentials: credentials.NewStaticCredentialsProvider(
			a.AccessKey,
			a.SecretKey,
			"",
		),
	}
	return s3.NewFromConfig(cfg), nil
}

func AllowWrite(file string) bool {

	// If the file is in the filterFiles map, return false
	var filterFiles = map[string]bool{
		"cors.json": true,
	}

	_, ok := filterFiles[file]
	return !ok
}

// UploadFile uploads the given file to the given bucket with the given key.
// It uploads the file with public-read permissions. If there is an error
// uploading the file, it logs an error and returns the error.
func (a *AwsManager) UploadFile(directory string, fileName string, file []byte, log *logger.Logger) (string, error) {
	log.Debug().Str("fileName", fileName).Msg("Uploading file to S3.")

	if fileName == "" {
		return "", fmt.Errorf("file name is required")
	}

	if !AllowWrite(fileName) {
		log.Warn().Str("fileName", fileName).Msg("File cannot be overriden")
		return "", nil
	}

	ctx := context.Background()
	objectKey := aws.String(fileName)

	client, err := a.GetClient()
	if err != nil {
		log.Error().Err(err).Msg("Error getting client")
		return "", err
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
		return "", err
	}

	return *objectKey, nil
}

// DeleteFile deletes the given file from the given bucket. It logs an error
// and returns the error if there is an error deleting the file.
func (a *AwsManager) DeleteFile(fileName string, log *logger.Logger) error {
	log.Info().Str("fileName", fileName).Msg("Deleting file from S3.")

	if fileName == "" {
		return fmt.Errorf("file name is required")
	}

	if !AllowWrite(fileName) {
		log.Warn().Str("fileName", fileName).Msg("File cannot be overriden")
		return fmt.Errorf("file cannot be overriden")
	}

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

func (a *AwsManager) DownloadFile(fileName string, log *logger.Logger) ([]byte, error) {

	if fileName == "" {
		return nil, fmt.Errorf("file name is required")
	}

	client, err := a.GetClient()
	if err != nil {
		log.Error().Err(err).Msg("Error getting client")
		return nil, err
	}

	ctx := context.Background()
	objectKey := aws.String(fileName)

	resp, err := client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(a.Bucket),
		Key:    objectKey,
	})
	if err != nil {
		log.Error().Err(err).Msg("Error downloading file from S3")
		return nil, err
	}

	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Error().Err(err).Msg("Error reading file from S3")
		return nil, err
	}

	return data, nil
}

func (a *AwsManager) ListFiles(log *logger.Logger) ([]string, error) {
	output := []string{}
	client, err := a.GetClient()
	if err != nil {
		log.Error().Err(err).Msg("Error getting client")
		return output, err
	}

	ctx := context.Background()
	resp, err := client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
		Bucket: aws.String(a.Bucket),
	})
	if err != nil {
		log.Error().Err(err).Msg("Error listing files from S3")
		return output, err
	}

	for _, obj := range resp.Contents {
		if *obj.Key == "cors.json" {
			log.Info().Msg("Skipping cors.json from listed files")
			continue
		}
		output = append(output, *obj.Key)
	}

	return output, nil
}

// GetFileInfo gets the file information from S3 for the given file name.
//
// It returns a FileInfo object containing the file's name, size, last modified
// time, and content type. If there is an error, it logs the error and returns
// it.
func (a *AwsManager) GetFileInfo(fileName string, log *logger.Logger) (*models.FileInfo, error) {

	if fileName == "" {
		return nil, fmt.Errorf("file name is required")
	}

	client, err := a.GetClient()
	if err != nil {
		log.Error().Err(err).Str("fileName", fileName).Msg("Error getting client")
		return nil, err
	}

	ctx := context.Background()
	resp, err := client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(a.Bucket),
		Key:    aws.String(fileName),
	})
	if err != nil {
		log.Error().Err(err).Str("fileName", fileName).Msg("Error getting file info from S3")
		return nil, err
	}

	fileInfo := &FileInfo{
		Name:         fileName,
		Size:         *resp.ContentLength,
		LastModified: *resp.LastModified,
		ContentType:  *resp.ContentType,
	}

	return fileInfo, err
}
