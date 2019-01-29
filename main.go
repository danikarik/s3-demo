package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

var (
	region     string
	apiKey     string
	secretKey  string
	bucketName string
	folderName string
)

func init() {
	errs := make([]error, 0)
	if region = os.Getenv("AWS_REGION"); region == "" {
		errs = append(errs, errors.New("AWS_REGION is not configured"))
	}
	if apiKey = os.Getenv("AWS_ACCESS_KEY_ID"); apiKey == "" {
		errs = append(errs, errors.New("AWS_ACCESS_KEY_ID is not configured"))
	}
	if secretKey = os.Getenv("AWS_SECRET_ACCESS_KEY"); secretKey == "" {
		errs = append(errs, errors.New("AWS_SECRET_ACCESS_KEY is not configured"))
	}
	if bucketName = os.Getenv("S3_BUCKET_NAME"); bucketName == "" {
		errs = append(errs, errors.New("S3_BUCKET_NAME is not configured"))
	}
	if folderName = os.Getenv("S3_FOLDER_NAME"); folderName == "" {
		errs = append(errs, errors.New("S3_FOLDER_NAME is not configured"))
	}
	if len(errs) > 0 {
		for _, err := range errs {
			fmt.Println(err.Error())
		}
		os.Exit(1)
	}
}

func main() {
	filename := ""
	flag.StringVar(&filename, "file", "", "filename to upload")
	flag.Parse()
	if filename == "" {
		exitErrorf("filename must be specified")
	}
	file, err := os.OpenFile(filename, os.O_RDWR|os.O_RDONLY, 0666)
	if err != nil {
		exitErrorf("could not open a file: %v", err)
	}
	defer file.Close()
	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String(region),
		Credentials: credentials.NewStaticCredentials(apiKey, secretKey, ""),
	})
	if err != nil {
		exitErrorf("could not create session: %v", err)
	}
	svc := s3.New(sess)
	uploader := s3manager.NewUploaderWithClient(svc, func(u *s3manager.Uploader) {
		u.PartSize = 64 * 1024 * 1024 // 64MB per part
	})
	_, path := filepath.Split(filename)
	output, err := uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(filepath.Join(folderName, path)),
		Body:   file,
	})
	if err != nil {
		exitErrorf("unable to upload %q to %q, %v", filename, bucketName, err)
	}
	fmt.Printf("successfully uploaded %q to %q\n", filename, bucketName)
	fmt.Printf("download url %s\n", output.Location)
}

func exitErrorf(msg string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, msg+"\n", args...)
	os.Exit(1)
}
