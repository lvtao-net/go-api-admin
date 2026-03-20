package storage

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/lvtao/go-gin-api-admin/internal/service"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// StorageService 存储服务接口
type StorageService interface {
	Upload(file *multipart.FileHeader, path string) (string, error)
	Download(path string) ([]byte, error)
	Delete(path string) error
	GetURL(path string) (string, error)
}

// LocalStorage 本地存储
type LocalStorage struct {
	basePath string
}

// NewLocalStorage 创建本地存储
func NewLocalStorage() *LocalStorage {
	settingService := service.NewSettingService()
	uploadPath := settingService.Get("upload.path")
	if uploadPath == "" {
		uploadPath = "./uploads"
	}
	return &LocalStorage{
		basePath: uploadPath,
	}
}

// Upload 上传文件
func (l *LocalStorage) Upload(file *multipart.FileHeader, path string) (string, error) {
	// 确保目录存在
	dir := filepath.Dir(filepath.Join(l.basePath, path))
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("failed to create directory: %w", err)
	}

	// 打开上传的文件
	src, err := file.Open()
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer src.Close()

	// 创建目标文件
	dst, err := os.Create(filepath.Join(l.basePath, path))
	if err != nil {
		return "", fmt.Errorf("failed to create file: %w", err)
	}
	defer dst.Close()

	// 复制内容
	if _, err := io.Copy(dst, src); err != nil {
		return "", fmt.Errorf("failed to copy file: %w", err)
	}

	return path, nil
}

// Download 下载文件
func (l *LocalStorage) Download(path string) ([]byte, error) {
	data, err := os.ReadFile(filepath.Join(l.basePath, path))
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}
	return data, nil
}

// Delete 删除文件
func (l *LocalStorage) Delete(path string) error {
	fullPath := filepath.Join(l.basePath, path)
	if err := os.Remove(fullPath); err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}
	return nil
}

// GetURL 获取文件URL
func (l *LocalStorage) GetURL(path string) (string, error) {
	// 返回相对路径，前端会拼接服务器地址
	return "/api/files/" + path, nil
}

// S3Storage S3存储
type S3Storage struct {
	client     *minio.Client
	bucket     string
	basePath   string
	endpoint   string
	region     string
}

// NewS3Storage 创建S3存储
func NewS3Storage() (*S3Storage, error) {
	settingService := service.NewSettingService()
	storageCfg := settingService.GetMap("storage.s3")
	if storageCfg == nil {
		storageCfg = make(map[string]interface{})
	}

	endpoint, _ := storageCfg["endpoint"].(string)
	accessKey, _ := storageCfg["accessKey"].(string)
	secretKey, _ := storageCfg["secretKey"].(string)
	bucket, _ := storageCfg["bucket"].(string)
	region, _ := storageCfg["region"].(string)
	if region == "" {
		region = "us-east-1"
	}

	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: endpoint != "localhost",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create S3 client: %w", err)
	}

	// 检查并创建 bucket
	exists, err := client.BucketExists(context.Background(), bucket)
	if err != nil {
		return nil, fmt.Errorf("failed to check bucket: %w", err)
	}
	if !exists {
		if err := client.MakeBucket(context.Background(), bucket, minio.MakeBucketOptions{
			Region: region,
		}); err != nil {
			return nil, fmt.Errorf("failed to create bucket: %w", err)
		}
	}

	uploadPath := settingService.Get("upload.path")
	if uploadPath == "" {
		uploadPath = "uploads"
	}

	return &S3Storage{
		client:   client,
		bucket:   bucket,
		basePath: uploadPath,
		endpoint: endpoint,
		region:   region,
	}, nil
}

// Upload 上传文件
func (s *S3Storage) Upload(file *multipart.FileHeader, path string) (string, error) {
	src, err := file.Open()
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer src.Close()

	// 读取文件内容
	data, err := io.ReadAll(src)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	// 上传到 S3
	objectPath := filepath.Join(s.basePath, path)
	_, err = s.client.PutObject(
		context.Background(),
		s.bucket,
		objectPath,
		bytes.NewReader(data),
		int64(len(data)),
		minio.PutObjectOptions{
			ContentType: file.Header.Get("Content-Type"),
		},
	)
	if err != nil {
		return "", fmt.Errorf("failed to upload to S3: %w", err)
	}

	return objectPath, nil
}

// Download 下载文件
func (s *S3Storage) Download(path string) ([]byte, error) {
	objectPath := filepath.Join(s.basePath, path)
	obj, err := s.client.GetObject(context.Background(), s.bucket, objectPath, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to download from S3: %w", err)
	}
	defer obj.Close()

	return io.ReadAll(obj)
}

// Delete 删除文件
func (s *S3Storage) Delete(path string) error {
	objectPath := filepath.Join(s.basePath, path)
	return s.client.RemoveObject(context.Background(), s.bucket, objectPath, minio.RemoveObjectOptions{})
}

// GetURL 获取文件URL
func (s *S3Storage) GetURL(path string) (string, error) {
	objectPath := filepath.Join(s.basePath, path)
	// 生成预签名 URL，有效期 24 小时
	reqParams := make(url.Values)
	presignedURL, err := s.client.PresignedGetObject(context.Background(), s.bucket, objectPath, 24*60*60, reqParams)
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}
	return presignedURL.String(), nil
}

// GetStorageService 获取存储服务实例
func GetStorageService() (StorageService, error) {
	settingService := service.NewSettingService()
	storageType := settingService.Get("storage.type")
	if storageType == "" {
		storageType = "local"
	}

	switch storageType {
	case "s3":
		return NewS3Storage()
	default:
		return NewLocalStorage(), nil
	}
}

// 生成缩略图（需要使用图像处理库如 github.com/disintegration/imaging）
// 这里提供接口，具体实现可根据需求添加
type ThumbnailGenerator interface {
	GenerateThumbnail(inputPath, outputPath string, width, height int) error
}

// ThumbnailConfig 缩略图配置
type ThumbnailConfig struct {
	Width  int `json:"width"`
	Height int `json:"height"`
}

// ImageProcessor 图片处理服务
type ImageProcessor struct {
	thumbnails map[string][]ThumbnailConfig
}

// NewImageProcessor 创建图片处理器
func NewImageProcessor() *ImageProcessor {
	return &ImageProcessor{
		thumbnails: map[string][]ThumbnailConfig{
			"small":  {{Width: 100, Height: 100}},
			"medium": {{Width: 300, Height: 300}},
			"large":  {{Width: 800, Height: 800}},
		},
	}
}

// IsImage 检查是否为图片
func IsImage(filename string) bool {
	ext := filepath.Ext(filename)
	imageExts := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".gif":  true,
		".webp": true,
		".bmp":  true,
	}
	return imageExts[strings.ToLower(ext)]
}

// GetThumbnailPath 获取缩略图路径
func GetThumbnailPath(path, size string) string {
	dir := filepath.Dir(path)
	ext := filepath.Ext(path)
	name := strings.TrimSuffix(filepath.Base(path), ext)
	return filepath.Join(dir, name+"_"+size+ext)
}
