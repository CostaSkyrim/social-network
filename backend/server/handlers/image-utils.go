package handlers

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"social-network/backend/config"
	"social-network/backend/global"
)

// SaveUploadedImage parses a multipart "image" field from the request, validates
// it against the configured limits (max size, allowed types), saves it to the
// uploads directory, and returns a URL-relative path like "images/{uuid}.jpg".
func SaveUploadedImage(r *http.Request) (string, error) {
	if err := parseMultipartForm(r); err != nil {
		return "", err
	}

	cfg := config.GetConfig()
	if cfg == nil {
		return "", fmt.Errorf("server configuration error")
	}
	allowedTypes := cfg.Handlers.Image.FileTypes
	if len(allowedTypes) == 0 {
		allowedTypes = []string{"image/jpeg", "image/png", "image/gif"}
	}

	maxSize, _ := imageMaxSize()

	file, header, err := r.FormFile("image")
	if err != nil {
		return "", fmt.Errorf("missing image file: %w", err)
	}
	defer file.Close()

	if header.Size > maxSize {
		return "", fmt.Errorf("file too large (max %d bytes)", maxSize)
	}

	ext, ok := extensionForType(header.Header.Get("Content-Type"), allowedTypes)
	if !ok {
		return "", fmt.Errorf("unsupported image type")
	}

	dir := global.GetImagePath(cfg.Handlers.Image.PathPrefix)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("failed to create uploads dir: %w", err)
	}

	filename := generateUUID() + ext
	dst, err := os.Create(filepath.Join(dir, filename))
	if err != nil {
		return "", fmt.Errorf("failed to create image file: %w", err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		return "", fmt.Errorf("failed to write image file: %w", err)
	}

	return "images/" + filename, nil
}

// SaveMultipartImage is a thin wrapper for handlers that returns an HTTP error
// response and a success bool, so callers don't duplicate the error handling.
func SaveMultipartImage(w http.ResponseWriter, r *http.Request) (string, bool) {
	path, err := SaveUploadedImage(r)
	if err != nil {
		RespondError(w, http.StatusBadRequest, err.Error())
		return "", false
	}
	return path, true
}

// isMultipart reports whether the request body is multipart/form-data.
func isMultipart(r *http.Request) bool {
	ct := r.Header.Get("Content-Type")
	return strings.HasPrefix(ct, "multipart/form-data")
}

// imageMaxSize returns the configured maximum upload size in bytes.
func imageMaxSize() (int64, error) {
	cfg := config.GetConfig()
	if cfg == nil {
		return 20 * 1024 * 1024, fmt.Errorf("server configuration error")
	}
	maxSize, err := global.ParseSize(cfg.Handlers.Image.MaxSize)
	if err != nil || maxSize <= 0 {
		return 20 * 1024 * 1024, nil // default 20MB
	}
	return maxSize, nil
}

// parseMultipartForm parses the request body as multipart/form-data once, using
// the configured image size limit as the memory threshold.
func parseMultipartForm(r *http.Request) error {
	if r.MultipartForm != nil {
		return nil
	}
	maxSize, _ := imageMaxSize()
	if err := r.ParseMultipartForm(maxSize + 1<<20); err != nil {
		return fmt.Errorf("failed to parse multipart form: %w", err)
	}
	return nil
}

// multipartImage saves the optional "image" file from an already-parsed
// multipart request. It returns (nil, true) when no image was provided,
// (path, true) on success, and (nil, false) after writing an error response.
func multipartImage(w http.ResponseWriter, r *http.Request) (*string, bool) {
	if r.MultipartForm == nil {
		return nil, true
	}
	if _, ok := r.MultipartForm.File["image"]; !ok {
		return nil, true
	}
	path, err := SaveUploadedImage(r)
	if err != nil {
		RespondError(w, http.StatusBadRequest, err.Error())
		return nil, false
	}
	return &path, true
}

// removeImageRequested reports whether the multipart request asks to clear an
// existing image (remove_image=1/true).
func removeImageRequested(r *http.Request) bool {
	v := r.FormValue("remove_image")
	return v == "1" || strings.EqualFold(v, "true")
}

func extensionForType(contentType string, allowedTypes []string) (string, bool) {
	// Normalize the content type (e.g. "image/jpeg; charset=binary" -> "image/jpeg")
	if idx := strings.IndexByte(contentType, ';'); idx != -1 {
		contentType = strings.TrimSpace(contentType[:idx])
	}

	var ext string
	switch contentType {
	case "image/jpeg", "image/jpg":
		ext = ".jpg"
	case "image/png":
		ext = ".png"
	case "image/gif":
		ext = ".gif"
	default:
		return "", false
	}

	if !containsString(allowedTypes, contentType) {
		return "", false
	}

	return ext, true
}

func containsString(list []string, s string) bool {
	for _, item := range list {
		if strings.EqualFold(item, s) {
			return true
		}
	}
	return false
}
