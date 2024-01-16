package database

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

func MinioClient() (*minio.Client, string) {
	loadErr := godotenv.Load()
	if loadErr != nil {
		log.Fatal("error loading file .env")
	}

	ssl, _ := strconv.ParseBool(os.Getenv("MINIO_SSL"))
	endpoint := os.Getenv("MINIO_ENDPOINT")
	accessKeyID := os.Getenv("MINIO_ACCESS_KEY")
	secretAccessKey := os.Getenv("MINIO_SECRET_KEY")
	bucketName := os.Getenv("MINIO_BUCKET")
	useSSL := ssl

	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: useSSL,
	})

	if err != nil {
		log.Fatal("error initializing Minio client: ", err)
	}

	return minioClient, bucketName
}
