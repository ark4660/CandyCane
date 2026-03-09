package storage

import (
	"io"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"mime/multipart"
	"net/http"

	"bytes"
	"log"
	"fmt"
)

func GetTranscriptUpload(file []byte, key string, s3Client *s3.Client) error {
	log.Println("log 1")
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("file", "audio.mp3")
	log.Printf("input to ExtractAudio: %d bytes / %d MB", len(file), len(file)/1024/1024)
	audioBytes, err := ExtractAudio(file)
	if err != nil {
		return err
	}
	log.Printf("audio size: %d MB", len(audioBytes)/1024/1024)
	log.Printf("audio size bytes: %d", len(audioBytes))
	reader := bytes.NewReader(audioBytes)
	io.Copy(part, reader)
	writer.WriteField("model", "whisper-1")
	writer.WriteField("language", "th")
	writer.Close()
	log.Println("log 2")
	req, err := http.NewRequest("POST", "https://api.openai.com/v1/audio/transcriptions", &body)
	if err != nil {
		return err
		log.Println(err)
	}
	req.Header.Set("Authorization", "Bearer "+ os.Getenv("OPENAI_API_KEY"))
	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
		return err

	}
	defer resp.Body.Close()
	log.Println("log 3")
	if resp.StatusCode != 200 {
        responsebody, _ := io.ReadAll(resp.Body)
        log.Println(resp.StatusCode)
        return fmt.Errorf("OpenAI API failed: %s", string(responsebody))
    }

	responsebody, err := io.ReadAll(resp.Body)
	if err != nil {
    	panic(err)
	}
 	if len(responsebody) == 0 {
        return fmt.Errorf("Empty response from OpenAI")
    }
	readCloser := io.NopCloser(bytes.NewReader(responsebody))
	err = UploadTranscriptJson(readCloser, key, s3Client)
	log.Println("log 4")
	if err != nil {
		log.Println(err)
		return err
	} else {
		return nil
	}
}
