package imgutil

import (
	"image"
	"image/color"
	"math"
	"sync"
)

func ConvertToGray(img image.Image) image.Image {
	bounds := img.Bounds()
	newImg := image.NewRGBA(bounds)

	for y := 0; y < bounds.Max.Y; y++ {
		for x := 0; x < bounds.Max.X; x++ {
			luminosity := pixelToGray(img.At(x, y).RGBA())
			pixel := color.Gray{Y: uint8(luminosity)}
			newImg.Set(x, y, pixel)
		}
	}

	return newImg
}

func ConvertToGrayArray(img image.Image) [][]float64 {
	switch it := img.(type) {
	case *image.YCbCr:
		return pixelToGrayYCbCR(it)
	case *image.RGBA:
		return pixelToGrayRGBA(it)
	default:
		return pixelToGrayDefault(it)
	}
}

func pixelToGrayYCbCR(img *image.YCbCr) [][]float64 {
	bounds := img.Bounds()
	w, h := bounds.Max.X-bounds.Min.X, bounds.Max.Y-bounds.Min.Y
	pixels := make([][]float64, h)

	for y := range pixels {
		pixels[y] = make([]float64, w)
		for x := range pixels[y] {
			pixels[y][x] = pixelToGray(img.YCbCrAt(x, y).RGBA())
		}
	}

	return pixels
}

func pixelToGrayRGBA(img *image.RGBA) [][]float64 {
	bounds := img.Bounds()
	w, h := bounds.Max.X-bounds.Min.X, bounds.Max.Y-bounds.Min.Y
	pixels := make([][]float64, h)

	for y := range pixels {
		pixels[y] = make([]float64, w)
		for x := range pixels[y] {
			pixels[y][x] = pixelToGray(img.At(x, y).RGBA())
		}
	}

	return pixels
}

func pixelToGrayDefault(img image.Image) [][]float64 {
	bounds := img.Bounds()
	w, h := bounds.Max.X-bounds.Min.X, bounds.Max.Y-bounds.Min.Y
	pixels := make([][]float64, h)

	for y := range pixels {
		pixels[y] = make([]float64, w)
		for x := range pixels[y] {
			pixels[y][x] = pixelToGray(img.At(x, y).RGBA())
		}
	}

	return pixels
}

func pixelToGray(r, g, b, a uint32) float64 {
	return 0.299*float64(r/257) + 0.587*float64(g/257) + 0.114*float64(b/256)
}

// DCT1D function returns result of DCT-II.
// DCT type II, unscaled. Algorithm by Byeong Gi Lee, 1984.
func DCT1D(input []float64) []float64 {
	temp := make([]float64, len(input))
	forwardTransform(input, temp, len(input))
	return input
}

func forwardTransform(input, temp []float64, Len int) {
	if Len == 1 {
		return
	}

	halfLen := Len / 2
	for i := 0; i < halfLen; i++ {
		x, y := input[i], input[Len-1-i]
		temp[i] = x + y
		temp[i+halfLen] = (x - y) / (math.Cos((float64(i)+0.5)*math.Pi/float64(Len)) * 2)
	}
	forwardTransform(temp, input, halfLen)
	forwardTransform(temp[halfLen:], input, halfLen)
	for i := 0; i < halfLen-1; i++ {
		input[i*2+0] = temp[i]
		input[i*2+1] = temp[i+halfLen] + temp[i+halfLen+1]
	}
	input[Len-2], input[Len-1] = temp[halfLen-1], temp[Len-1]
}

// DCT2D function returns a  result of DCT2D by using the seperable property.
func DCT2D(input [][]float64, w int, h int) [][]float64 {
	output := make([][]float64, h)
	for i := range output {
		output[i] = make([]float64, w)
	}

	wg := new(sync.WaitGroup)
	for i := 0; i < h; i++ {
		wg.Add(1)
		go func(i int) {
			cols := DCT1D(input[i])
			output[i] = cols
			wg.Done()
		}(i)
	}

	wg.Wait()
	for i := 0; i < w; i++ {
		wg.Add(1)
		in := make([]float64, h)
		go func(i int) {
			for j := 0; j < h; j++ {
				in[j] = output[j][i]
			}
			rows := DCT1D(in)
			for j := 0; j < len(rows); j++ {
				output[j][i] = rows[j]
			}
			wg.Done()
		}(i)
	}

	wg.Wait()
	return output
}
