package imageutil

import (
	"image"
	"image/color"
)

func ConvertToGray(img image.Image) image.Image {
	bounds := img.Bounds()
	newImg := image.NewRGBA(bounds)

	for y := 0; y < bounds.Max.Y; y++ {
		for x := 0; x < bounds.Max.X; x++ {
			colorPixel := img.At(x, y)
			r, g, b, _ := colorPixel.RGBA()
			luminosity := 0.299*float64(r>>8) + 0.587*float64(g>>8) + 0.114*float64(b>>8)
			pixel := color.Gray{Y: uint8(luminosity)}
			newImg.Set(x, y, pixel)
		}
	}

	return newImg
}

func ConvertToGrayArray(img image.Image) [][]float64 {
	bounds := img.Bounds()
	w, h := bounds.Max.X-bounds.Min.X, bounds.Max.Y-bounds.Min.Y
	pixels := make([][]float64, h)

	for y := range pixels {
		pixels[y] = make([]float64, w)
		for x := range pixels[y] {
			r, g, b, _ := img.At(x, y).RGBA()
			lum := 0.299*float64(r>>8) + 0.587*float64(g>>8) + 0.114*float64(b>>8)
			pixels[y][x] = lum
		}
	}

	return pixels
}
