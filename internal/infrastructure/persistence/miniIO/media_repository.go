package minioS3

import (
	"bytes"
	"context"
	"fmt"

	"github.com/pedrodcsjostrom/opencm/internal/domain/media"
	"github.com/pedrodcsjostrom/opencm/internal/infrastructure/config"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

type S3Client struct {
	cfg    *config.ObjectStoreConfig
	client *s3.S3
	bucket string
}

func NewS3Client(cfg *config.ObjectStoreConfig) (*S3Client, error) {

	sess, err := session.NewSession(&aws.Config{
		Region:           aws.String(cfg.Region),
		Credentials:      credentials.NewStaticCredentials(cfg.AccessKey, cfg.SecretAccessKey, ""),
		Endpoint:         aws.String(cfg.Endpoint),
		DisableSSL:       aws.Bool(!cfg.UseSSL),
		S3ForcePathStyle: aws.Bool(cfg.S3ForcePathStyle),
	})

	if err != nil {
		return nil, err
	}

	return &S3Client{
		cfg:    cfg,
		client: s3.New(sess),
		bucket: cfg.Bucket,
	}, nil
}

func (c *S3Client) UploadFile(ctx context.Context, projectID, postID, fileName string, data []byte, metadata *media.MetaData)  error {
	key := c.getKey(projectID, postID, fileName)
	_, err := c.client.PutObjectWithContext(ctx, &s3.PutObjectInput{
		Bucket: aws.String(c.cfg.Bucket),
		Key:    aws.String(key),
		Body:   bytes.NewReader(data),
	})
	if err != nil {
		return  err
	}
	return nil
}

func (c *S3Client) GetFile(ctx context.Context, projectID, postID, fileName string) ([]byte, error) {
	key := c.getKey(projectID, postID, fileName)
	result, err := c.client.GetObjectWithContext(ctx, &s3.GetObjectInput{
		Bucket: aws.String(c.cfg.Bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, err
	}
	defer result.Body.Close()
	
	buf := new(bytes.Buffer)
	buf.ReadFrom(result.Body)
	return buf.Bytes(), nil
}

func (c *S3Client) getKey(projectID, postID, fileName string) string {
	return fmt.Sprintf("project-%s/post-%s/%s", projectID, postID, fileName)
}

// // returns true if filename starts with "th-"
// func (c *S3Client) isThumbnail(fileName string) bool {
// 	return fileName[:3] == "th-"
// }