package storage

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

const maxClockPhotoSizeBytes int64 = 10 * 1024 * 1024

var allowedImageContentTypes = map[string]string{
	"image/jpeg": ".jpg",
	"image/jpg":  ".jpg",
	"image/png":  ".png",
	"image/webp": ".webp",
}

type SupabaseStorage struct {
	baseURL        string
	serviceRoleKey string
	bucket         string
	httpClient     *http.Client

	ensureBucketOnce sync.Once
	ensureBucketErr  error
}

func NewSupabaseStorage(baseURL, serviceRoleKey, bucket string) *SupabaseStorage {
	return &SupabaseStorage{
		baseURL:        strings.TrimRight(baseURL, "/"),
		serviceRoleKey: strings.TrimSpace(serviceRoleKey),
		bucket:         strings.TrimSpace(bucket),
		httpClient: &http.Client{
			Timeout: 20 * time.Second,
		},
	}
}

func (s *SupabaseStorage) UploadClockPhoto(ctx context.Context, userID int64, fileHeader *multipart.FileHeader) (string, string, error) {
	if strings.TrimSpace(s.baseURL) == "" || strings.TrimSpace(s.serviceRoleKey) == "" || strings.TrimSpace(s.bucket) == "" {
		return "", "", fmt.Errorf("configuração do Supabase Storage incompleta")
	}

	if fileHeader == nil {
		return "", "", fmt.Errorf("arquivo de foto não enviado")
	}

	if fileHeader.Size <= 0 {
		return "", "", fmt.Errorf("arquivo de foto vazio")
	}

	if fileHeader.Size > maxClockPhotoSizeBytes {
		return "", "", fmt.Errorf("arquivo excede o limite de 10MB")
	}

	s.ensureBucketOnce.Do(func() {
		s.ensureBucketErr = s.ensureBucket(ctx)
	})
	if s.ensureBucketErr != nil {
		return "", "", s.ensureBucketErr
	}

	file, err := fileHeader.Open()
	if err != nil {
		return "", "", fmt.Errorf("erro ao abrir arquivo de foto: %w", err)
	}
	defer file.Close()

	buffer := make([]byte, 512)
	n, _ := io.ReadFull(file, buffer)
	detectedContentType := http.DetectContentType(buffer[:n])
	contentType := normalizeContentType(detectedContentType)
	ext, ok := allowedImageContentTypes[contentType]
	if !ok {
		return "", "", fmt.Errorf("formato de foto inválido, use JPEG, PNG ou WEBP")
	}

	objectPath := s.buildObjectPath(userID, fileHeader.Filename, ext)
	body := io.MultiReader(bytes.NewReader(buffer[:n]), file)

	uploadURL := fmt.Sprintf("%s/storage/v1/object/%s/%s", s.baseURL, s.bucket, objectPath)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, uploadURL, body)
	if err != nil {
		return "", "", fmt.Errorf("erro ao criar requisição de upload: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+s.serviceRoleKey)
	req.Header.Set("apikey", s.serviceRoleKey)
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("x-upsert", "false")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", "", fmt.Errorf("erro ao enviar foto para o storage: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		return "", "", fmt.Errorf("falha ao salvar foto no storage (status %d): %s", resp.StatusCode, strings.TrimSpace(string(respBody)))
	}

	publicURL := fmt.Sprintf("%s/storage/v1/object/public/%s/%s", s.baseURL, s.bucket, objectPath)
	return objectPath, publicURL, nil
}

func (s *SupabaseStorage) ensureBucket(ctx context.Context) error {
	bucketURL := fmt.Sprintf("%s/storage/v1/bucket/%s", s.baseURL, s.bucket)

	getReq, err := http.NewRequestWithContext(ctx, http.MethodGet, bucketURL, nil)
	if err != nil {
		return fmt.Errorf("erro ao criar requisição para validar bucket: %w", err)
	}
	getReq.Header.Set("Authorization", "Bearer "+s.serviceRoleKey)
	getReq.Header.Set("apikey", s.serviceRoleKey)

	getResp, err := s.httpClient.Do(getReq)
	if err != nil {
		return fmt.Errorf("erro ao validar bucket no storage: %w", err)
	}
	defer getResp.Body.Close()

	if getResp.StatusCode == http.StatusOK {
		return nil
	}

	if getResp.StatusCode != http.StatusNotFound {
		respBody, _ := io.ReadAll(io.LimitReader(getResp.Body, 2048))
		return fmt.Errorf("erro ao consultar bucket (status %d): %s", getResp.StatusCode, strings.TrimSpace(string(respBody)))
	}

	createBody, err := json.Marshal(map[string]interface{}{
		"id":     s.bucket,
		"name":   s.bucket,
		"public": true,
	})
	if err != nil {
		return fmt.Errorf("erro ao montar payload para criação do bucket: %w", err)
	}

	createURL := fmt.Sprintf("%s/storage/v1/bucket", s.baseURL)
	createReq, err := http.NewRequestWithContext(ctx, http.MethodPost, createURL, bytes.NewReader(createBody))
	if err != nil {
		return fmt.Errorf("erro ao criar requisição de criação de bucket: %w", err)
	}
	createReq.Header.Set("Authorization", "Bearer "+s.serviceRoleKey)
	createReq.Header.Set("apikey", s.serviceRoleKey)
	createReq.Header.Set("Content-Type", "application/json")

	createResp, err := s.httpClient.Do(createReq)
	if err != nil {
		return fmt.Errorf("erro ao criar bucket no storage: %w", err)
	}
	defer createResp.Body.Close()

	if createResp.StatusCode < 200 || createResp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(io.LimitReader(createResp.Body, 2048))
		return fmt.Errorf("erro ao criar bucket (status %d): %s", createResp.StatusCode, strings.TrimSpace(string(respBody)))
	}

	return nil
}

func (s *SupabaseStorage) buildObjectPath(userID int64, originalName, fallbackExt string) string {
	ext := strings.ToLower(filepath.Ext(originalName))
	if ext == "" {
		ext = fallbackExt
	}

	now := time.Now().UTC()
	uniquePart := fmt.Sprintf("%d_%d", now.UnixNano(), userID)
	return fmt.Sprintf("users/%d/%s%s", userID, uniquePart, ext)
}

func normalizeContentType(contentType string) string {
	contentType = strings.TrimSpace(strings.ToLower(contentType))
	if idx := strings.Index(contentType, ";"); idx >= 0 {
		contentType = strings.TrimSpace(contentType[:idx])
	}
	return contentType
}
