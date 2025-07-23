package uploader

import (
	"crypto/md5"
	"encoding/hex"
	"io"
	"mime"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

// S3Uploader implements the Driver base on the S3 of AWS.
type S3Uploader struct {
	client *s3.S3
	bucket string
}

// NewS3Uploader returns a new s3 uploader
func NewS3Uploader(region, endpoint, accessKey, secretKey, bucketName string) (Driver, error) {
	cfg := aws.NewConfig().WithCredentials(credentials.NewStaticCredentials(accessKey, secretKey, ""))
	s, err := session.NewSession(cfg)
	if err != nil {
		return nil, err
	}

	return &S3Uploader{
		client: s3.New(s, cfg.WithRegion(region), cfg.WithEndpoint(endpoint)),
		bucket: bucketName,
	}, nil
}

// ListObjects returns some remote objects
func (u *S3Uploader) ListObjects(prefix string) ([]Object, error) {
	marker := ""
	objects := make([]Object, 0)
	for {
		input := &s3.ListObjectsInput{
			Bucket: aws.String(u.bucket),
			Prefix: aws.String(prefix),
			Marker: aws.String(marker),
		}
		objectsResult, err := u.client.ListObjects(input)
		if err != nil {
			return nil, err
		}

		for _, obj := range objectsResult.Contents {
			fObj := Object{
				Key:  aws.StringValue(obj.Key),
				ETag: strings.Trim(aws.StringValue(obj.ETag), `"`),
			}

			objects = append(objects, fObj)
		}

		if aws.BoolValue(objectsResult.IsTruncated) {
			marker = aws.StringValue(objectsResult.NextMarker)
		} else {
			break
		}
	}

	return objects, nil
}

// Upload uploads the local file to the object
func (u *S3Uploader) Upload(objectKey, filePath string) (err error) {
	bodyReader, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer bodyReader.Close()

	_, err = u.client.PutObject(&s3.PutObjectInput{
		Body:        bodyReader,
		Bucket:      aws.String(u.bucket),
		Key:         aws.String(objectKey),
		ContentType: aws.String(DetectContentType(filePath)),
	})
	return
}

// Delete deletes the object
func (u *S3Uploader) Delete(object string) (err error) {
	_, err = u.client.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(u.bucket),
		Key:    aws.String(object),
	})
	return
}

// DetectContentType returns the file content-type
func DetectContentType(filepath string) string {
	mimeType := mime.TypeByExtension(path.Ext(filepath))
	if mimeType != "" {
		return mimeType
	}

	fileData, err := os.ReadFile(filepath)
	if err != nil {
		return ""
	}

	return http.DetectContentType(fileData)
}

// MD5Hex returns the file md5 hash hex
func MD5Hex(filepath string) string {
	f, err := os.Open(filepath)
	if err != nil {
		return ""
	}
	defer f.Close()

	md5hash := md5.New()
	if _, err := io.Copy(md5hash, f); err != nil {
		return ""
	}

	return hex.EncodeToString(md5hash.Sum(nil)[:])
}
