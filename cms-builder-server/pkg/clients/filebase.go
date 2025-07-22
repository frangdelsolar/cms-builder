package clients

import (
	"bytes"
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"

	fileTypes "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/file/types"
	loggerTypes "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/logger/types"
)

type FilebaseManager struct {
	Bucket    string
	Endpoint  string
	Region    string
	AccessKey string
	SecretKey string
}

func (a *FilebaseManager) GetClient() (*s3.S3, error) {
	//// create a configuration
	s3Config := aws.Config{
		Credentials:      credentials.NewStaticCredentials(a.AccessKey, a.SecretKey, ""),
		Endpoint:         aws.String(a.Endpoint),
		Region:           aws.String(a.Region),
		S3ForcePathStyle: aws.Bool(true),
	}

	goSession, err := session.NewSessionWithOptions(session.Options{
		Config:  s3Config,
		Profile: "filebase",
	})

	// check if the session was created correctly.
	if err != nil {
		return nil, err
	}

	// create a s3 client session
	s3Client := s3.New(goSession)

	return s3Client, nil
}

func (a *FilebaseManager) UploadFile(filePath string, file []byte, log *loggerTypes.Logger) (string, error) {
	log.Debug().Str("fileName", filePath).Msg("Uploading file to Filebase.")

	if filePath == "" {
		return "", fmt.Errorf("file name is required")
	}

	client, err := a.GetClient()
	if err != nil {
		log.Error().Err(err).Msg("Error getting client")
		return "", err
	}

	objectKey := aws.String(filePath)

	_, err = client.PutObject(&s3.PutObjectInput{
		Body:   bytes.NewReader(file),
		Bucket: aws.String(a.Bucket),
		Key:    objectKey,
	})
	if err != nil {
		log.Error().Err(err).Msg("Error uploading file to Filebase")
		return "", err
	}

	return *objectKey, nil
}

func (a *FilebaseManager) DeleteFile(fileName string, log *loggerTypes.Logger) error {
	log.Info().Str("fileName", fileName).Msg("Deleting file from Filebase.")

	if fileName == "" {
		return fmt.Errorf("file name is required")
	}

	if !AllowWrite(fileName) {
		log.Warn().Str("fileName", fileName).Msg("File cannot be overriden")
		return fmt.Errorf("file cannot be overriden")
	}

	objectKey := aws.String(fileName)

	client, err := a.GetClient()
	if err != nil {
		log.Error().Err(err).Msg("Error getting client")
		return err
	}

	_, err = client.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(a.Bucket),
		Key:    objectKey,
	})
	if err != nil {
		log.Error().Err(err).Msg("Error deleting file from Filebase")
		return err
	}

	return nil
}

func (a *FilebaseManager) DownloadFile(fileName string, log *loggerTypes.Logger) ([]byte, error) {

	if fileName == "" {
		return nil, fmt.Errorf("file name is required")
	}

	client, err := a.GetClient()
	if err != nil {
		log.Error().Err(err).Msg("Error getting client")
		return nil, err
	}

	objectKey := aws.String(fileName)

	resp, err := client.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(a.Bucket),
		Key:    objectKey,
	})
	if err != nil {
		log.Error().Err(err).Msg("Error downloading file from Filebase")
		return nil, err
	}

	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Error().Err(err).Msg("Error reading file from Filebase")
		return nil, err
	}

	return data, nil
}

func (a *FilebaseManager) ListFiles(log *loggerTypes.Logger) ([]string, error) {
	output := []string{}
	client, err := a.GetClient()
	if err != nil {
		log.Error().Err(err).Msg("Error getting client")
		return output, err
	}

	resp, err := client.ListObjectsV2(&s3.ListObjectsV2Input{
		Bucket: aws.String(a.Bucket),
	})
	if err != nil {
		log.Error().Err(err).Msg("Error listing files from Filebase")
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

func (a *FilebaseManager) GetFileInfo(fileName string, log *loggerTypes.Logger) (*fileTypes.FileInfo, error) {

	if fileName == "" {
		return nil, fmt.Errorf("file name is required")
	}

	client, err := a.GetClient()
	if err != nil {
		log.Error().Err(err).Str("fileName", fileName).Msg("Error getting client")
		return nil, err
	}

	resp, err := client.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(a.Bucket),
		Key:    aws.String(fileName),
	})
	if err != nil {
		log.Error().Err(err).Str("fileName", fileName).Msg("Error getting file info from Filebase")
		return nil, err
	}

	fileInfo := &fileTypes.FileInfo{
		Name:         fileName,
		Size:         *resp.ContentLength,
		LastModified: *resp.LastModified,
		ContentType:  *resp.ContentType,
	}

	return fileInfo, err
}
