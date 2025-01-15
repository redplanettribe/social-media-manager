package media

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"image/jpeg"
	_ "image/png"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"golang.org/x/image/draw"
)

var (
	ErrUnsupportedMediaType = errors.New("unsupported media type")
	allowedImageFormats     = map[string]bool{"jpg": true, "jpeg": true, "png": true}
	allowedVideoFormats     = map[string]bool{"mp4": true, "mov": true}
	ThumbnailFormat         = "jpeg"
)

type MediaProcessor interface {
	Analyze(data []byte) (*MediaInfo, error)
	GetThumbnail(data []byte) (*[]byte, error)
}

type MediaInfo struct {
	Type   MediaType
	Format string
	Width  int
	Height int
	Length int // for videos
	Size   int // in bytes
}

func GetProcessor(filename string) (MediaProcessor, error) {
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
		Size:   len(data),
	}, nil
}

func (a *ImageAnalyzer) GetThumbnail(imgData []byte) (*[]byte, error) {
    // 1. Decode source image
    src, format, err := image.Decode(bytes.NewReader(imgData))
    if err != nil {
        return nil, fmt.Errorf("failed to decode image: %w", err)
    }

    // 2. Convert to RGBA to ensure JPEG compatibility
    bounds := src.Bounds()
    rgbaImg := image.NewRGBA(bounds)
    draw.Draw(rgbaImg, bounds, src, bounds.Min, draw.Src)

    // 3. Create thumbnail size RGBA
    thumbSize := image.Rect(0, 0, 100, 100)
    thumbnail := image.NewRGBA(thumbSize)

    // 4. Scale down preserving aspect ratio
    srcAspect := float64(bounds.Dx()) / float64(bounds.Dy())
    dstAspect := float64(thumbSize.Dx()) / float64(thumbSize.Dy())

    var r image.Rectangle
    if srcAspect > dstAspect {
        // Source is wider
        h := int(float64(thumbSize.Dx()) / srcAspect)
        r = image.Rect(0, (100-h)/2, 100, (100+h)/2)
    } else {
        // Source is taller
        w := int(float64(thumbSize.Dy()) * srcAspect)
        r = image.Rect((100-w)/2, 0, (100+w)/2, 100)
    }

    // 5. Use high quality scaling
    draw.CatmullRom.Scale(thumbnail, r, rgbaImg, bounds, draw.Over, nil)

    // 6. Encode to JPEG with specific quality
    var out bytes.Buffer
    if err := jpeg.Encode(&out, thumbnail, &jpeg.Options{
        Quality: 85,
    }); err != nil {
        return nil, fmt.Errorf("failed to encode JPEG thumbnail: %w", err)
    }

    result := out.Bytes()
    fmt.Printf("Thumbnail created from %s image, size: %d bytes\n", format, len(result))
    return &result, nil
}

type VideoAnalyzer struct {
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

func (a *VideoAnalyzer) Analyze(video []byte) (*MediaInfo, error) {
	// Create a temporary file to store the video data
	tmpFile, err := os.CreateTemp("", "video-*"+"."+a.ext)
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tmpFile.Name())

	// Write video data to temp file
	if _, err := tmpFile.Write(video); err != nil {
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
		Size:   len(video),
	}, nil
}

func (a *VideoAnalyzer) GetThumbnail(video []byte) (*[]byte, error) {
	tmpFile, err := os.CreateTemp("", "video-*"+"."+a.ext)
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.Write(video); err != nil {
		return nil, fmt.Errorf("failed to write temp file: %w", err)
	}
	if err := tmpFile.Close(); err != nil {
		return nil, fmt.Errorf("failed to close temp file: %w", err)
	}

	// Prepare ffmpeg command
	cmd := exec.Command("ffmpeg",
		"-i", tmpFile.Name(),
		"-ss", "00:00:01.000",
		"-vframes", "1",
		"-f", "image2",
		"-")

	var output bytes.Buffer
	cmd.Stdout = &output
	cmd.Stderr = nil

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("ffmpeg failed: %w", err)
	}

	result := output.Bytes()
	return &result, nil
}
