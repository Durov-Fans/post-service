package photo

import (
	"fmt"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"net/http"
	"post-service/internal/lib/uploaders"
)

func ProcessPhoto(r *http.Request, fieldName string, client *s3.Client, tgHash, postUUID string) (string, error) {
	file, header, err := r.FormFile(fieldName)
	if err != nil || header.Size == 0 {
		return "", nil
	}
	defer file.Close()

	data := make([]byte, header.Size)
	_, err = file.Read(data)
	if err != nil {
		return "", fmt.Errorf("read error: %w", err)
	}

	errUpload, url := uploaders.UploadPost(client, tgHash, fieldName, data, postUUID)
	if errUpload != nil {
		return "", fmt.Errorf("upload error: %w", errUpload)
	}
	return url, nil
}
