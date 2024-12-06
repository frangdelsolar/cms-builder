package builder

import (
	"image"
	"image/jpeg"
	"os"
)

func SaveImage(img image.Image, path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	if err := jpeg.Encode(file, img, nil); err != nil {
		return err
	}

	return nil
}

// func ResizeImage(img image.Image, width, height int) (image.Image, error) {

// }
