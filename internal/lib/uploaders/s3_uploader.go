package uploaders

import (
	"bytes"
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/joho/godotenv"
	"log"
	"os"
	"strings"
)

func InitAWS() *s3.Client {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Не загрузился .env file")
	}

	customResolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		return aws.Endpoint{
			URL:           os.Getenv("S3_ENDPOINT"),
			SigningRegion: os.Getenv("AWS_REGION"),
		}, nil
	})

	cfg := aws.Config{
		Credentials:                 credentials.NewStaticCredentialsProvider(os.Getenv("AWS_ACCESS_KEY_ID"), os.Getenv("AWS_SECRET_ACCESS_KEY"), ""),
		Region:                      os.Getenv("AWS_REGION"),
		EndpointResolverWithOptions: customResolver,
	}

	return s3.NewFromConfig(cfg)
}

func ListBuckets(client *s3.Client) {
	resp, err := client.ListBuckets(context.TODO(), &s3.ListBucketsInput{})
	if err != nil {
		log.Fatalf("Ошибка при получении списка бакетов: %v", err)
	}

	for _, b := range resp.Buckets {
		log.Printf("Бакет: %s\n", *b.Name)
	}
}

func UploadFile(client *s3.Client, userHash string, fileName string, fileBytes []byte) (error, string) {
	body := bytes.NewReader(fileBytes)

	var path string
	if fileName == "avatar" {
		path = "avatar"
	} else if fileName == "banner" {
		path = "banner"
	}

	if !strings.HasSuffix(fileName, ".png") {
		fileName += ".png"
	}

	key := userHash + "/" + path + "/" + fileName

	_, err := client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket:      aws.String(os.Getenv("AWS_BUCKET")),
		Key:         &key,
		Body:        body,
		ContentType: aws.String("image/png"),
		ACL:         "public-read",
	})
	if err != nil {
		return fmt.Errorf("Ошибка загрузки: %w", err), "0"
	}

	log.Println("Успешная загрузка")

	endpoint := os.Getenv("S3_ENDPOINT")
	bucket := os.Getenv("AWS_BUCKET")

	// Убери https:// и / в конце
	endpoint = strings.TrimSuffix(strings.TrimPrefix(endpoint, "https://"), "/")

	// 🔗 Публичная ссылка: https://s3.timeweb.com/<bucket>/<key>
	publicURL := fmt.Sprintf("https://%s/%s/%s", endpoint, bucket, key)

	return nil, publicURL
}

func UploadPost(client *s3.Client, userHash string, fileName string, fileBytes []byte, postPath string) (error, string) {
	body := bytes.NewReader(fileBytes)

	if !strings.HasSuffix(fileName, ".png") {
		fileName += ".png"
	}

	key := userHash + "/" + "posts" + "/" + postPath + "/" + fileName

	_, err := client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket:      aws.String(os.Getenv("AWS_BUCKET")),
		Key:         &key,
		Body:        body,
		ContentType: aws.String("image/png"),
		ACL:         "public-read",
	})
	if err != nil {
		return fmt.Errorf("Ошибка загрузки: %w", err), "0"
	}

	log.Println("Успешная загрузка")

	endpoint := os.Getenv("S3_ENDPOINT")
	bucket := os.Getenv("AWS_BUCKET")

	// Убери https:// и / в конце
	endpoint = strings.TrimSuffix(strings.TrimPrefix(endpoint, "https://"), "/")

	// 🔗 Публичная ссылка: https://s3.timeweb.com/<bucket>/<key>
	publicURL := fmt.Sprintf("https://%s/%s/%s", endpoint, bucket, key)

	return nil, publicURL
}
