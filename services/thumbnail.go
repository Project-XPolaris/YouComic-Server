package services

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"github.com/allentom/harukap/thumbnail"
	"github.com/nfnt/resize"
	"github.com/projectxpolaris/youcomic/config"
	thumbnail2 "github.com/projectxpolaris/youcomic/thumbnail"
	"image"
	_ "image/gif"
	"image/jpeg"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
)

const MaxThumbnailGeneratorQueue = 10

var DefaultThumbnailService = NewThumbnailService(100)

type ThumbnailTaskOption struct {
	Input   string
	Output  string
	ErrChan chan error
}
type ThumbnailServiceStatus struct {
	sync.Mutex
	Total      int64
	InQueue    int64
	MaxQueue   int64
	InProgress int64
}
type ThumbnailService struct {
	MaxTask  int
	Resource chan ThumbnailTaskOption
	Status   ThumbnailServiceStatus
}

func NewThumbnailService(maxTask int) *ThumbnailService {
	resChan := make(chan ThumbnailTaskOption, MaxThumbnailGeneratorQueue)
	useChan := make(chan struct{}, maxTask)
	service := &ThumbnailService{
		Resource: resChan,
		MaxTask:  maxTask,
		Status: ThumbnailServiceStatus{
			Total:      0,
			InQueue:    0,
			MaxQueue:   MaxThumbnailGeneratorQueue,
			InProgress: 0,
		},
	}
	go func() {
		for {
			useChan <- struct{}{}
			option := <-resChan
			service.Status.Lock()
			service.Status.InQueue = int64(len(service.Resource))
			service.Status.Unlock()
			go func() {
				defer func() {
					<-useChan
				}()
				service.Status.Lock()
				service.Status.InProgress += 1
				service.Status.Unlock()
				_, err := GenerateCoverThumbnail(option.Input, option.Output)
				service.Status.Lock()
				service.Status.InProgress -= 1
				service.Status.Total += 1
				service.Status.Unlock()
				if err != nil {
					option.ErrChan <- err
					return
				}
				option.ErrChan <- nil
			}()

		}
	}()
	return service
}

func (s *ThumbnailService) GetQueueStatus() *ThumbnailServiceStatus {
	s.Status.Lock()
	defer s.Status.Unlock()
	return &s.Status
}

func GetImageProcess() ImageProcessEngine {
	var generator ImageProcessEngine
	switch config.Instance.Thumbnail.Type {
	case "vips":
		generator = NewVipsThumbnailEngine(config.Instance.Thumbnail.Target)
	case "thumbnailservice":
		generator = &ThumbnailServiceEngine{}
	default:
		generator = &DefaultThumbnailsEngine{}
	}
	return generator
}

//generate thumbnail image
func GenerateCoverThumbnail(coverImageFilePath string, storePath string) (string, error) {
	var err error
	defer func() {
		if r := recover(); r != nil {
			if er, ok := r.(error); ok {
				err = er
			} else {
				err = errors.New("unknown panic error")
			}
		}
	}()
	// setup image decoder
	fileExt := filepath.Ext(coverImageFilePath)
	// mkdir
	err = os.MkdirAll(storePath, os.ModePerm)
	if err != nil {
		return "", err
	}
	thumbnailImagePath := filepath.Join(storePath, fmt.Sprintf("cover_thumbnail%s", fileExt))

	abs, err := filepath.Abs(thumbnailImagePath)
	generator := GetImageProcess()
	if err != nil {
		return "", err
	}
	output, err := generator.GenerateThumbnail(coverImageFilePath, abs, 480)
	if err != nil {
		generator = &DefaultThumbnailsEngine{}
		return generator.GenerateThumbnail(coverImageFilePath, abs, 480)
	}
	return output, err
}
func ResizeImageWithSizeCap(input string, targetSizeCap int64) ([]byte, error) {
	sourceImageFile, err := os.Open(input)
	if err != nil {
		return nil, err
	}
	stats, err := sourceImageFile.Stat()
	if err != nil {
		return nil, err
	}
	fileSize := stats.Size()
	if fileSize < targetSizeCap {
		if err != nil {
			return nil, err
		}
		buffer := bufio.NewReader(sourceImageFile)
		imageBytes := make([]byte, stats.Size())

		_, err := buffer.Read(imageBytes)
		if err != nil {
			return nil, err
		}
		return imageBytes, nil
	}

	imageConf, _, err := image.DecodeConfig(sourceImageFile)
	reduceRatio := float64(targetSizeCap) / float64(fileSize)
	reduceWidth := int(float64(imageConf.Width) * reduceRatio)
	reduceHeight := int(float64(imageConf.Height) * reduceRatio)
	process := GetImageProcess()
	output, err := process.Resize(input, reduceWidth, reduceHeight)
	if err != nil {
		return nil, err
	}
	return output, nil
}

type ImageProcessEngine interface {
	GenerateThumbnail(input string, output string, maxWidth int) (path string, err error)
	Resize(input string, width int, height int) ([]byte, error)
}

type VipsThumbnailEngine struct {
	Target string
}

func NewVipsThumbnailEngine(target string) *VipsThumbnailEngine {
	return &VipsThumbnailEngine{Target: target}
}

func (e *VipsThumbnailEngine) GenerateThumbnail(input string, output string, maxWidth int) (string, error) {
	cmd := exec.Command(e.Target, fmt.Sprintf("--size=%dx", maxWidth), input, "-o", output)
	err := cmd.Run()
	if err != nil {
		return "", err
	}
	return output, nil
}
func (e *VipsThumbnailEngine) Resize(input string, width int, height int) ([]byte, error) {
	engine := DefaultThumbnailsEngine{}
	return engine.Resize(input, width, height)
}

type DefaultThumbnailsEngine struct {
}

func (e *DefaultThumbnailsEngine) loadImageFromFile(input string) (image.Image, error) {
	thumbnailImageFile, err := os.Open(input)
	defer thumbnailImageFile.Close()
	if err != nil {
		return nil, err
	}
	thumbnailImage, _, err := image.Decode(thumbnailImageFile)
	if err != nil {
		return nil, err
	}
	return thumbnailImage, nil
}
func (e *DefaultThumbnailsEngine) GenerateThumbnail(input string, output string, maxWidth int) (path string, err error) {
	thumbnailImage, err := e.loadImageFromFile(input)
	if err != nil {
		return "", err
	}
	// make thumbnail
	resizeImage := resize.Thumbnail(uint(maxWidth), 480, thumbnailImage, resize.Lanczos3)

	// mkdir
	outputImage, err := os.Create(output)
	if err != nil {
		return "", err
	}

	defer outputImage.Close()

	// save result
	err = jpeg.Encode(outputImage, resizeImage, nil)
	if err != nil {
		return "", err
	}
	return output, nil
}

func (e *DefaultThumbnailsEngine) Resize(input string, width int, height int) ([]byte, error) {
	sourceImage, err := e.loadImageFromFile(input)
	if err != nil {
		return nil, err
	}

	resizeImage := resize.Resize(uint(width), uint(height), sourceImage, resize.Lanczos3)
	buf := new(bytes.Buffer)
	err = jpeg.Encode(buf, resizeImage, nil)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

type ThumbnailServiceEngine struct {
}

func (t *ThumbnailServiceEngine) GenerateThumbnail(input string, output string, maxWidth int) (path string, err error) {
	err = thumbnail2.DefaultThumbnailServicePlugin.Client.Generate(input, output, thumbnail.ThumbnailOption{
		MaxWidth: maxWidth,
		Mode:     "width",
	})
	if err != nil {
		return "", err
	}
	return output, nil
}

func (t *ThumbnailServiceEngine) Resize(input string, width int, height int) ([]byte, error) {
	out, err := thumbnail2.DefaultThumbnailServicePlugin.Client.Resize(input, thumbnail.ThumbnailOption{
		MaxWidth:  width,
		MaxHeight: height,
		Mode:      "resize",
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}
