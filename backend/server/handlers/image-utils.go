package handlers

import (
	"bytes"
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"social-network/backend/config"
	"social-network/backend/global"
)

// defaultImageTypes and defaultImageMaxSize are the fallbacks used when the
// config omits the corresponding settings.
var defaultImageTypes = []string{"image/jpeg", "image/png", "image/gif"}

const defaultImageMaxSize = 20 * 1024 * 1024

// maxImageDimension caps the longest edge of any resized image so huge uploads
// are downscaled on the server before storage and serving.
const maxImageDimension = 2048

// allowedTypesFor returns the configured allowed MIME types, applying the
// default set when none are configured.
func allowedTypesFor(cfg *config.Config) []string {
	if cfg != nil && len(cfg.Handlers.Image.FileTypes) > 0 {
		return cfg.Handlers.Image.FileTypes
	}
	return defaultImageTypes
}

// SaveUploadedImage parses a multipart "image" field from the request, validates
// it against the configured limits (max size, allowed types — verified by
// sniffing the actual bytes, not the client header), downscales it server-side
// to a bounded dimension, saves it to the uploads directory, and returns a
// URL-relative path like "images/{uuid}.jpg".
func SaveUploadedImage(r *http.Request) (string, error) {
	if err := parseMultipartForm(r); err != nil {
		return "", err
	}

	cfg := config.GetConfig()
	if cfg == nil {
		return "", fmt.Errorf("server configuration error")
	}
	allowedTypes := allowedTypesFor(cfg)
	maxSize := imageMaxSize()

	file, header, err := r.FormFile("image")
	if err != nil {
		return "", fmt.Errorf("missing image file: %w", err)
	}
	defer file.Close()

	if header.Size > maxSize {
		return "", fmt.Errorf("file too large (max %d bytes)", maxSize)
	}

	// Read the whole file into memory so we can both sniff its type and decode
	// it. Uploads are already capped at maxSize so this is bounded.
	data, err := io.ReadAll(io.LimitReader(file, maxSize+1))
	if err != nil {
		return "", fmt.Errorf("failed to read image: %w", err)
	}
	if int64(len(data)) > maxSize {
		return "", fmt.Errorf("file too large (max %d bytes)", maxSize)
	}

	sniffed := sniffImageType(data)
	if !containsString(allowedTypes, sniffed) {
		return "", fmt.Errorf("unsupported image type")
	}

	ext := extensionForFormat(sniffed)

	// Downscale/encode so stored images stay bounded regardless of source size.
	processed, err := processImage(data, sniffed)
	if err != nil {
		return "", err
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

	if _, err := dst.Write(processed); err != nil {
		return "", fmt.Errorf("failed to write image file: %w", err)
	}

	return "images/" + filename, nil
}

// SaveMultipartImage writes an error response and returns (path, ok) for
// handlers that only ever expect a single required image.
func SaveMultipartImage(w http.ResponseWriter, r *http.Request) (string, bool) {
	path, err := SaveUploadedImage(r)
	if err != nil {
		RespondError(w, http.StatusBadRequest, err.Error())
		return "", false
	}
	return path, true
}

// DeleteImageFile removes an uploaded image from disk given its URL-relative
// path (e.g. "images/{uuid}.jpg"). It ignores paths that aren't under the
// uploads directory and never errors on a missing file.
func DeleteImageFile(relativePath *string) {
	if relativePath == nil || *relativePath == "" {
		return
	}
	p := *relativePath
	if !strings.HasPrefix(p, "images/") {
		return
	}
	dir := global.GetImagePath(config.GetConfig().Handlers.Image.PathPrefix)
	// Prevent path traversal: only allow a plain filename after the prefix.
	filename := strings.TrimPrefix(p, "images/")
	if filename == "" || strings.ContainsAny(filename, `/\`) {
		return
	}
	_ = os.Remove(filepath.Join(dir, filename))
}

// resolveImageUpdate computes the resulting image path for an edit operation
// with a "keep / replace / remove" tri-state:
//   - a newly uploaded image wins,
//   - otherwise remove_image=true clears it,
//   - otherwise the existing path is kept.
func resolveImageUpdate(old *string, newPath *string, r *http.Request) *string {
	switch {
	case newPath != nil:
		return newPath
	case removeImageRequested(r):
		return nil
	default:
		return old
	}
}

// isMultipart reports whether the request body is multipart/form-data.
func isMultipart(r *http.Request) bool {
	ct := r.Header.Get("Content-Type")
	return strings.HasPrefix(ct, "multipart/form-data")
}

// imageMaxSize returns the configured maximum upload size in bytes.
func imageMaxSize() int64 {
	cfg := config.GetConfig()
	if cfg == nil {
		return defaultImageMaxSize
	}
	maxSize, err := global.ParseSize(cfg.Handlers.Image.MaxSize)
	if err != nil || maxSize <= 0 {
		return defaultImageMaxSize
	}
	return maxSize
}

// parseMultipartForm parses the request body as multipart/form-data once, using
// the configured image size limit as the memory threshold.
func parseMultipartForm(r *http.Request) error {
	if r.MultipartForm != nil {
		return nil
	}
	if err := r.ParseMultipartForm(imageMaxSize() + 1<<20); err != nil {
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

// sniffImageType inspects a file's leading bytes and returns its canonical MIME
// type using the standard library's magic-byte detection. The result is checked
// against the allowed-types list by the caller, so an unrecognized or spoofed
// type is rejected regardless of the client-supplied Content-Type header.
func sniffImageType(data []byte) string {
	return http.DetectContentType(data)
}

// extensionForFormat maps a canonical image MIME type to its file extension.
func extensionForFormat(contentType string) string {
	switch contentType {
	case "image/jpeg", "image/jpg":
		return ".jpg"
	case "image/png":
		return ".png"
	case "image/gif":
		return ".gif"
	case "image/webp":
		return ".webp"
	default:
		return ""
	}
}

// extensionForType maps a content type to an extension, checking it against an
// allow-list. Kept for the OAuth avatar-download path which validates against a
// provider-reported header rather than a user upload.
func extensionForType(contentType string, allowedTypes []string) (string, bool) {
	if idx := strings.IndexByte(contentType, ';'); idx != -1 {
		contentType = strings.TrimSpace(contentType[:idx])
	}
	if !containsString(allowedTypes, contentType) {
		return "", false
	}
	if ext := extensionForFormat(contentType); ext != "" {
		return ext, true
	}
	return "", false
}

// processImage decodes, downscales (when above maxImageDimension), and
// re-encodes the image data. Re-encoding also normalizes and strips any
// extraneous bytes/payloads. GIF is re-encoded losslessly (no resizing, since
// animated frames can't be resized with the stdlib without breaking animation).
func processImage(data []byte, contentType string) ([]byte, error) {
	var (
		img image.Image
		err error
	)

	switch contentType {
	case "image/jpeg":
		img, err = jpeg.Decode(bytes.NewReader(data))
	case "image/png":
		img, err = png.Decode(bytes.NewReader(data))
	case "image/gif":
		// Re-encode GIF as-is to normalize; don't resize animated GIFs.
		return reencodeGIF(data)
	case "image/webp":
		// The stdlib can't decode WebP; pass through the original bytes.
		return data, nil
	default:
		return nil, fmt.Errorf("unsupported image type")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	img = downscale(img, maxImageDimension)

	var buf bytes.Buffer
	switch contentType {
	case "image/jpeg":
		err = jpeg.Encode(&buf, img, &jpeg.Options{Quality: 85})
	case "image/png":
		err = png.Encode(&buf, img)
	default:
		return nil, fmt.Errorf("unsupported image type")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to encode image: %w", err)
	}

	return buf.Bytes(), nil
}

// downscale scales img down so its longest edge is at most maxEdge. Images
// already under the limit are returned unchanged.
func downscale(img image.Image, maxEdge int) image.Image {
	b := img.Bounds()
	w, h := b.Dx(), b.Dy()
	if w <= maxEdge && h <= maxEdge {
		return img
	}

	var nw, nh int
	if w >= h {
		nw = maxEdge
		nh = h * maxEdge / w
	} else {
		nh = maxEdge
		nw = w * maxEdge / h
	}
	if nw < 1 {
		nw = 1
	}
	if nh < 1 {
		nh = 1
	}

	// Bilinear via a simple nearest-free approach using image/draw scaling.
	dst := image.NewRGBA(image.Rect(0, 0, nw, nh))
	// Nearest-neighbour scaling keeps it simple and dependency-free.
	for y := 0; y < nh; y++ {
		sy := y * h / nh
		for x := 0; x < nw; x++ {
			sx := x * w / nw
			dst.Set(x, y, img.At(b.Min.X+sx, b.Min.Y+sy))
		}
	}
	return dst
}

// reencodeGIF decodes and re-encodes a GIF to normalize it (preserving animation
// via image/gif's All decode).
func reencodeGIF(data []byte) ([]byte, error) {
	g, err := gif.DecodeAll(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("failed to decode gif: %w", err)
	}
	var buf bytes.Buffer
	if err := gif.EncodeAll(&buf, g); err != nil {
		return nil, fmt.Errorf("failed to encode gif: %w", err)
	}
	return buf.Bytes(), nil
}

func containsString(list []string, s string) bool {
	for _, item := range list {
		if strings.EqualFold(item, s) {
			return true
		}
	}
	return false
}
