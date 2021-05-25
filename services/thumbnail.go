package services

import (
	"fmt"
	"github.com/nfnt/resize"
	"image"
	"image/jpeg"
	"image/png"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

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
	return output,nil
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
	resizeImage := resize.Thumbnail(uint(maxWidth), 0, thumbnailImage, resize.Lanczos3)

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
