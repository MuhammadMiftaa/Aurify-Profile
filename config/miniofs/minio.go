package miniofs

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"net/url"
	"strings"
	"sync"
	"time"

	"refina-profile/config/env"
	"refina-profile/config/log"
	constant "refina-profile/internal/utils/data"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// FileValidationConfig holds validation rules
type FileValidationConfig struct {
	AllowedExtensions []string
	MaxFileSize       int64 // in bytes
	MinFileSize       int64 // in bytes
}

// UploadRequest represents file upload request
type UploadRequest struct {
	Base64Data string
	Prefix     string // prefix for filename, will be combined with timestamp and extension
	BucketName string
	Validation *FileValidationConfig
}

// UploadResponse represents upload result
type UploadResponse struct {
	BucketName string
	ObjectName string
	Size       int64
	URL        string
	Ext        string
	ETag       string
}

type MinIOConfig struct {
	Host           string
	AccessKey      string
	SecretKey      string
	UseSSL         bool
	MaxConnections int
	ConnectTimeout time.Duration
	RequestTimeout time.Duration
	BucketName     string
}

// Global MinIO manager - singleton pattern seperti database/redis
type MinIOManager struct {
	client      *minio.Client
	config      MinIOConfig
	mu          sync.RWMutex
	isReady     bool
	bucketCache map[string]bool // cache untuk bucket existence check
}

var (
	MinioClient *MinIOManager
	once        sync.Once
)

// Init initializes global MinIO manager - dipanggil sekali di main.go
func SetupMinio(cfg env.Minio) *MinIOManager {
	once.Do(func() {
		minioCfg := MinIOConfig{
			Host:           cfg.Host,
			AccessKey:      cfg.AccessKey,
			SecretKey:      cfg.SecretKey,
			UseSSL:         cfg.UseSSL == 1,
			ConnectTimeout: 30 * time.Second,
			RequestTimeout: 60 * time.Second,
			BucketName:     cfg.BucketName,
		}

		var err error
		MinioClient, err = newMinIOManager(minioCfg)
		if err != nil {
			log.Log.Fatalf("Failed to initialize MinIO: %v", err)
		}
	})

	return MinioClient
}

// newMinIOManager creates new MinIO manager
func newMinIOManager(cfg MinIOConfig) (*MinIOManager, error) {
	client, err := minio.New(cfg.Host, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create minio client: %v", err)
	}

	manager := &MinIOManager{
		client:      client,
		config:      cfg,
		bucketCache: make(map[string]bool),
	}

	// Verify connection by checking bucket
	ctx, cancel := context.WithTimeout(context.Background(), cfg.ConnectTimeout)
	defer cancel()

	exists, err := client.BucketExists(ctx, cfg.BucketName)
	if err != nil {
		return nil, fmt.Errorf("failed to check bucket: %v", err)
	}

	if !exists {
		// Create bucket if not exists
		err = client.MakeBucket(ctx, cfg.BucketName, minio.MakeBucketOptions{})
		if err != nil {
			return nil, fmt.Errorf("failed to create bucket: %v", err)
		}
		log.Info(constant.LogMinioBucketCreated, map[string]any{"bucket": cfg.BucketName})
	}

	manager.bucketCache[cfg.BucketName] = true
	manager.isReady = true

	return manager, nil
}

// GetClient returns the MinIO client
func (m *MinIOManager) GetClient() *minio.Client {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.client
}

// GetBucketName returns the configured bucket name
func (m *MinIOManager) GetBucketName() string {
	return m.config.BucketName
}

// IsReady checks if MinIO is ready
func (m *MinIOManager) IsReady() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.isReady
}

// UploadBase64 uploads a base64 encoded file
func (m *MinIOManager) UploadBase64(ctx context.Context, req UploadRequest) (*UploadResponse, error) {
	if !m.IsReady() {
		return nil, fmt.Errorf("minio client is not ready")
	}

	// Decode base64
	data, ext, err := decodeBase64Image(req.Base64Data)
	if err != nil {
		return nil, fmt.Errorf("failed to decode base64: %v", err)
	}

	// Validate if rules provided
	if req.Validation != nil {
		if err := validateFile(data, ext, req.Validation); err != nil {
			return nil, err
		}
	}

	// Generate unique filename
	objectName := fmt.Sprintf("%s_%d%s", req.Prefix, time.Now().UnixNano(), ext)

	// Get content type
	contentType := getContentType(ext)

	// Upload
	bucketName := req.BucketName
	if bucketName == "" {
		bucketName = m.config.BucketName
	}

	info, err := m.client.PutObject(ctx, bucketName, objectName, bytes.NewReader(data), int64(len(data)), minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to upload: %v", err)
	}

	// Generate URL
	objectURL := m.GetObjectURL(bucketName, objectName)

	return &UploadResponse{
		BucketName: bucketName,
		ObjectName: objectName,
		Size:       info.Size,
		URL:        objectURL,
		Ext:        ext,
		ETag:       info.ETag,
	}, nil
}

// DeleteObject deletes an object from MinIO
func (m *MinIOManager) DeleteObject(ctx context.Context, bucketName, objectName string) error {
	if !m.IsReady() {
		return fmt.Errorf("minio client is not ready")
	}

	if bucketName == "" {
		bucketName = m.config.BucketName
	}

	err := m.client.RemoveObject(ctx, bucketName, objectName, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete object: %v", err)
	}

	return nil
}

// GetObjectURL generates URL for an object
func (m *MinIOManager) GetObjectURL(bucketName, objectName string) string {
	protocol := "http"
	if m.config.UseSSL {
		protocol = "https"
	}
	return fmt.Sprintf("%s://%s/%s/%s", protocol, m.config.Host, bucketName, objectName)
}

// ExtractObjectName extracts object name from URL
func (m *MinIOManager) ExtractObjectName(photoURL string) (string, error) {
	if photoURL == "" {
		return "", fmt.Errorf("empty photo URL")
	}

	u, err := url.Parse(photoURL)
	if err != nil {
		return "", fmt.Errorf("invalid URL: %v", err)
	}

	// Path format: /bucket/objectName
	parts := strings.SplitN(strings.TrimPrefix(u.Path, "/"), "/", 2)
	if len(parts) < 2 {
		return "", fmt.Errorf("invalid URL format")
	}

	return parts[1], nil
}

// Helper functions

func decodeBase64Image(base64Data string) ([]byte, string, error) {
	// Handle data URL format: data:image/png;base64,xxxxx
	if strings.HasPrefix(base64Data, "data:") {
		parts := strings.SplitN(base64Data, ",", 2)
		if len(parts) != 2 {
			return nil, "", fmt.Errorf("invalid data URL format")
		}

		// Extract extension from mime type
		mimeType := strings.TrimPrefix(parts[0], "data:")
		mimeType = strings.TrimSuffix(mimeType, ";base64")
		ext := mimeTypeToExt(mimeType)

		data, err := base64.StdEncoding.DecodeString(parts[1])
		if err != nil {
			return nil, "", err
		}

		return data, ext, nil
	}

	// Plain base64
	data, err := base64.StdEncoding.DecodeString(base64Data)
	if err != nil {
		return nil, "", err
	}

	// Detect extension from magic bytes
	ext := detectExtension(data)
	return data, ext, nil
}

func mimeTypeToExt(mimeType string) string {
	switch mimeType {
	case "image/jpeg", "image/jpg":
		return ".jpg"
	case "image/png":
		return ".png"
	case "image/gif":
		return ".gif"
	case "image/webp":
		return ".webp"
	default:
		return ".jpg" // default
	}
}

func detectExtension(data []byte) string {
	if len(data) < 4 {
		return ".jpg"
	}

	// PNG
	if data[0] == 0x89 && data[1] == 0x50 && data[2] == 0x4E && data[3] == 0x47 {
		return ".png"
	}
	// JPEG
	if data[0] == 0xFF && data[1] == 0xD8 {
		return ".jpg"
	}
	// GIF
	if data[0] == 0x47 && data[1] == 0x49 && data[2] == 0x46 {
		return ".gif"
	}
	// WEBP
	if len(data) >= 12 && string(data[8:12]) == "WEBP" {
		return ".webp"
	}

	return ".jpg"
}

func getContentType(ext string) string {
	switch ext {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".gif":
		return "image/gif"
	case ".webp":
		return "image/webp"
	default:
		return "application/octet-stream"
	}
}

func validateFile(data []byte, ext string, cfg *FileValidationConfig) error {
	// Check size
	size := int64(len(data))
	if cfg.MaxFileSize > 0 && size > cfg.MaxFileSize {
		return fmt.Errorf("file size %d exceeds maximum %d bytes", size, cfg.MaxFileSize)
	}
	if cfg.MinFileSize > 0 && size < cfg.MinFileSize {
		return fmt.Errorf("file size %d is below minimum %d bytes", size, cfg.MinFileSize)
	}

	// Check extension
	if len(cfg.AllowedExtensions) > 0 {
		allowed := false
		for _, e := range cfg.AllowedExtensions {
			if ext == e {
				allowed = true
				break
			}
		}
		if !allowed {
			return fmt.Errorf("extension %s is not allowed", ext)
		}
	}

	return nil
}
