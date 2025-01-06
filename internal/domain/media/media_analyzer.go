package media

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

var (
	ErrUnsupportedMediaType = errors.New("unsupported media type")
	allowedImageFormats     = map[string]bool{"jpg": true, "jpeg": true, "png": true}
	allowedVideoFormats     = map[string]bool{"mp4": true, "mov": true}
)

type MediaAnalyzer interface {
	Analyze(data []byte) (*MediaInfo, error)
}

type MediaInfo struct {
	Type   MediaType
	Format string
	Width  int
	Height int
	Length int // for videos
}

func GetAnalyzer(filename string) (MediaAnalyzer, error) {
	ext := strings.ToLower(filepath.Ext(filename))
	if ext == "" {
		return nil, ErrUnsupportedMediaType
	}
	ext = ext[1:] // remove dot

	switch {
	case allowedImageFormats[ext]:
		return &ImageAnalyzer{
			ext: ext,
		}, nil
	case allowedVideoFormats[ext]:
		return &VideoAnalyzer{
			ext: ext,
		}, nil
	default:
		return nil, ErrUnsupportedMediaType
	}
}

// IMPLEMENTATION

type ImageAnalyzer struct {
	ext string
}

func (a *ImageAnalyzer) Analyze(data []byte) (*MediaInfo, error) {
	img, format, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}

	return &MediaInfo{
		Type:   MediaTypeImage,
		Format: format,
		Width:  img.Width,
		Height: img.Height,
	}, nil
}

type VideoAnalyzer struct{
	ext string
}

type ffprobeOutput struct {
	Streams []struct {
		CodecType string `json:"codec_type"`
		Width     int    `json:"width,omitempty"`
		Height    int    `json:"height,omitempty"`
	} `json:"streams"`
	Format struct {
		Duration string `json:"duration"`
	} `json:"format"`
}

func (a *VideoAnalyzer) Analyze(data []byte) (*MediaInfo, error) {
	// Create a temporary file to store the video data
	tmpFile, err := os.CreateTemp("", "video-*"+"."+a.ext)
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tmpFile.Name()) 

	// Write video data to temp file
	if _, err := tmpFile.Write(data); err != nil {
		return nil, fmt.Errorf("failed to write temp file: %w", err)
	}
	if err := tmpFile.Close(); err != nil {
		return nil, fmt.Errorf("failed to close temp file: %w", err)
	}

	// Prepare ffprobe command
	cmd := exec.Command("ffprobe",
		"-v", "quiet",
		"-print_format", "json",
		"-show_format",
		"-show_streams",
		tmpFile.Name())

	var output bytes.Buffer
	cmd.Stdout = &output
	cmd.Stderr = nil

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("ffprobe failed: %w", err)
	}

	// Parse ffprobe output
	var ffdata ffprobeOutput
	if err := json.Unmarshal(output.Bytes(), &ffdata); err != nil {
		return nil, fmt.Errorf("failed to parse ffprobe output: %w", err)
	}

	// Find video stream and get dimensions
	var width, height int
	for _, stream := range ffdata.Streams {
		if stream.CodecType == "video" {
			width = stream.Width
			height = stream.Height
			break
		}
	}

	if width == 0 || height == 0 {
		return nil, fmt.Errorf("no video stream found or invalid dimensions")
	}

	// Parse duration to seconds
	var length float64
	fmt.Sscanf(ffdata.Format.Duration, "%f", &length)

	return &MediaInfo{
		Type:   MediaTypeVideo,
		Format: a.ext,
		Width:  width,
		Height: height,
		Length: int(length),
	}, nil
}