package utils

import (
	"fmt"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

var UploadBaseDir = "uploads"

func InitUploadDir() {
	for _, sub := range []string{"absensi", "pendaftaran", "sertifikat"} {
		_ = os.MkdirAll(filepath.Join(UploadBaseDir, sub), 0o755)
	}
}

func UploadBasePath() string {
	if p := os.Getenv("UPLOAD_PATH"); p != "" {
		return p
	}
	return UploadBaseDir
}

func SaveUploadedFile(c *gin.Context, field, subdir string, allowedExt map[string]bool, maxBytes int64) (string, error) {
	file, err := c.FormFile(field)
	if err != nil {
		return "", fmt.Errorf("file %s wajib diunggah", field)
	}
	if file.Size > maxBytes {
		return "", fmt.Errorf("ukuran file maksimal %d MB", maxBytes/(1024*1024))
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	if !allowedExt[ext] {
		return "", fmt.Errorf("ekstensi file tidak diizinkan: %s", ext)
	}

	base := UploadBasePath()
	dir := filepath.Join(base, subdir)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}

	name := fmt.Sprintf("%s_%d%s", uuid.New().String()[:8], time.Now().Unix(), ext)
	dest := filepath.Join(dir, name)

	if err := c.SaveUploadedFile(file, dest); err != nil {
		return "", err
	}

	return filepath.ToSlash(filepath.Join(subdir, name)), nil
}

func AllowedImageExt() map[string]bool {
	return map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".webp": true}
}

func AllowedPDFExt() map[string]bool {
	return map[string]bool{".pdf": true}
}

func AllowedDocExt() map[string]bool {
	return map[string]bool{".pdf": true, ".jpg": true, ".jpeg": true, ".png": true, ".webp": true}
}

const MaxImageSize = 5 * 1024 * 1024  // 5MB
const MaxPDFSize = 10 * 1024 * 1024   // 10MB

func OptionalFormFile(c *gin.Context, field string) (*multipart.FileHeader, error) {
	f, err := c.FormFile(field)
	if err != nil {
		return nil, nil
	}
	return f, nil
}
