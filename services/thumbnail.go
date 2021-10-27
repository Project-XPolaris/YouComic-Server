package services

import (
	"fmt"
	"github.com/allentom/youcomic-api/config"
	"github.com/nfnt/resize"
	"image"
	"image/jpeg"
	"image/png"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
)

const MaxThumbnailGeneratorQueue = 100000

var DefaultThumbnailService = NewThumbnailService(10)

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

//generate thumbnail image
func GenerateCoverThumbnail(coverImageFilePath string, storePath string) (string, error) {
	// setup image decoder
	fileExt := filepath.Ext(coverImageFilePath)
	// mkdir
	err := os.MkdirAll(storePath, os.ModePerm)
	if err != nil {
		return "", err
	}
	thumbnailImagePath := filepath.Join(storePath, fmt.Sprintf("cover_thumbnail%s", fileExt))

	var generator ThumbnailEngine
	if config.Config.Thumbnail.Type == "vips" {
		generator = NewVipsThumbnailEngine(config.Config.Thumbnail.Target)
	} else {
		generator = &DefaultThumbnailsEngine{}
	}
	abs, err := filepath.Abs(thumbnailImagePath)
	if err != nil {
		return "", err
	}
	output, err := generator.Generate(coverImageFilePath, abs, 480)
	if err != nil {
		generator = &DefaultThumbnailsEngine{}
		return generator.Generate(coverImageFilePath, abs, 480)
	}
	return output, err
}

type ThumbnailEngine interface {
	Generate(input string, output string, maxWidth int) (path string, err error)
}

type VipsThumbnailEngine struct {
	Target string
}

func NewVipsThumbnailEngine(target string) *VipsThumbnailEngine {
	return &VipsThumbnailEngine{Target: target}
}

func (e *VipsThumbnailEngine) Generate(input string, output string, maxWidth int) (string, error) {
	cmd := exec.Command(e.Target, fmt.Sprintf("--size=%dx", maxWidth), input, "-o", output)
	err := cmd.Run()
	if err != nil {
		return "", err
	}
	return output, nil
}

type DefaultThumbnailsEngine struct {
}

func (e *DefaultThumbnailsEngine) Generate(input string, output string, maxWidth int) (path string, err error) {
	fileExt := filepath.Ext(input)
	thumbnailImageFile, err := os.Open(input)
	if err != nil {
		return "", err
	}
	var thumbnailImage image.Image
	if strings.ToLower(fileExt) == ".png" {
		thumbnailImage, err = png.Decode(thumbnailImageFile)
	}
	if strings.ToLower(fileExt) == ".jpg" {
		thumbnailImage, err = jpeg.Decode(thumbnailImageFile)
	}
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

	defer thumbnailImageFile.Close()
	defer outputImage.Close()

	// save result
	err = jpeg.Encode(outputImage, resizeImage, nil)
	if err != nil {
		return "", err
	}
	return output, nil
}
