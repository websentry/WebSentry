package utils

import (
	"image"
	"image/png"
	"log"
	"math"
	"math/rand"
	"os"
	"path"
	"strings"

	"github.com/disintegration/imaging"
	"github.com/pkg/errors"

	"github.com/websentry/websentry/config"
)

const imageFilenameChar = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890"

var imageBasePath, imageThumbBasePath string

func init() {
	imageBasePath = path.Join(config.GetConfig().FileStoragePath, "sentry", "image", "orig")
	imageThumbBasePath = path.Join(config.GetConfig().FileStoragePath, "sentry", "image", "thumb")

	err := os.MkdirAll(imageBasePath, os.ModePerm)
	if err != nil {
		panic(err)
	}
	err = os.MkdirAll(imageThumbBasePath, os.ModePerm)
	if err != nil {
		panic(err)
	}
}

func RandStringBytes(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = imageFilenameChar[rand.Intn(len(imageFilenameChar))]
	}
	return string(b)
}

func ImageRandomFilename() string {
	var filename string

	for {
		filename = RandStringBytes(32)

		fullFilename := path.Join(imageThumbBasePath, filename+".jpg")
		_, err := os.Stat(fullFilename)
		if os.IsNotExist(err) {
			break
		}
	}

	return filename
}

func ImageCheckFilename(filename string) bool {
	for _, char := range filename {
		if !strings.Contains(imageFilenameChar, string(char)) {
			return false
		}
	}
	return true
}

// need check filename if the filename comes from user
func ImageGetFullPath(filename string, thumb bool) string {
	if thumb {
		return path.Join(imageThumbBasePath, filename+".jpg")
	} else {
		return path.Join(imageBasePath, filename+".png")
	}
}

func deleteFileAndIgnoreError(filename string) {
	err := os.Remove(filename)
	if err != nil && !os.IsNotExist(err) {
		err = errors.Wrapf(err, "File: %v", filename)
		log.Printf("deleteFileAndIgnoreError: \n%+v", err)
	}
}

// if failed, only log the error
func ImageDelete(filename string, keepThumb bool) {
	deleteFileAndIgnoreError(ImageGetFullPath(filename, false))
	if !keepThumb {
		deleteFileAndIgnoreError(ImageGetFullPath(filename, true))
	}
}

func pixelDifference(a uint32, b uint32) float64 {
	return math.Abs(float64(a)-float64(b)) / 65535.0
}

func ImageCompare(a image.Image, b image.Image) (float32, error) {
	if a.Bounds() != b.Bounds() {
		return 0, errors.New("images with different size")
	}

	bounds := a.Bounds()
	total := 0
	v := 0.0
	for i := bounds.Min.X; i < bounds.Max.X; i++ {
		for j := bounds.Min.Y; j < bounds.Max.Y; j++ {

			ar, ag, ab, _ := a.At(i, j).RGBA()
			br, bg, bb, _ := b.At(i, j).RGBA()
			v += pixelDifference(ar, br)
			v += pixelDifference(ag, bg)
			v += pixelDifference(ab, bb)

			total += 3
		}
	}
	return 1 - float32(v/float64(total)), nil
}

func ImageSave(image image.Image) (string, error) {
	filename := ImageRandomFilename()

	err := imaging.Save(image, ImageGetFullPath(filename, false), imaging.PNGCompressionLevel(png.BestCompression))
	err = errors.WithStack(err)
	if err != nil {
		return "", err
	}
	// thumb
	err = imaging.Save(image, ImageGetFullPath(filename, true), imaging.JPEGQuality(70))
	err = errors.WithStack(err)
	if err != nil {
		return "", err
	}

	return filename, nil
}
