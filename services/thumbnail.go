package services

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"github.com/allentom/harukap/plugins/thumbnail"
	"github.com/projectxpolaris/youcomic/plugin"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io/ioutil"
	"os"
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

// generate thumbnail image
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
	thumbnailImagePath := filepath.Join(storePath, fmt.Sprintf("cover_thumbnail%s", fileExt))
	coverSourceFile, err := os.Open(coverImageFilePath)
	if err != nil {
		return "", err
	}
	out, err := plugin.ThumbnailEngine.Resize(context.Background(), coverSourceFile, thumbnail.ThumbnailOption{
		MaxWidth:  480,
		MaxHeight: 480,
	})
	storage := plugin.GetDefaultStorage()
	err = storage.Upload(context.Background(), out, plugin.GetDefaultBucket(), thumbnailImagePath)
	if err != nil {
		return "", err
	}
	return thumbnailImagePath, nil
}

func DirectUploadCoverThumbnail(coverFilePath string) (string, error) {
	coverSourceFile, err := os.Open(coverFilePath)
	if err != nil {
		return "", err
	}
	storage := plugin.GetDefaultStorage()
	thumbnailImagePath := fmt.Sprintf("cover_thumbnail%s", filepath.Ext(coverFilePath))
	err = storage.Upload(context.Background(), coverSourceFile, plugin.GetDefaultBucket(), thumbnailImagePath)
	if err != nil {
		return "", err
	}
	return thumbnailImagePath, nil
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
	out, err := plugin.ThumbnailEngine.Resize(context.Background(), sourceImageFile, thumbnail.ThumbnailOption{
		MaxWidth:  reduceWidth,
		MaxHeight: reduceHeight,
	})
	if err != nil {
		return nil, err
	}

	output, err := ioutil.ReadAll(out)
	if err != nil {
		return nil, err
	}
	return output, nil
}
