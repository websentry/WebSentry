package utils

import (
	"math"
	"image"
	"path"
	"github.com/websentry/websentry/config"
	"time"
	"math/rand"
	"strconv"
	"os"
	"io/ioutil"
	"errors"
)

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
	return 1 - float32(v / float64(total)), nil
}

func ImageSave(b []byte) string {

	// generate file name
	basePath := path.Join(config.GetFileStoragePath(), "sentry", "image")
	stime := time.Now().Format("20060102150405")
	filename := ""
	fullFilename := ""
	for {
		i := rand.Intn(200)

		filename = stime + "-" + strconv.Itoa(i) + ".png"
		fullFilename = path.Join(basePath, filename)
		_, err := os.Stat(fullFilename)
		if os.IsNotExist(err) {
			break
		}
	}

	// save
	ioutil.WriteFile(fullFilename, b, 0644)
	return filename
}
