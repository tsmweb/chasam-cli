package imageutil

import (
	"fmt"
	"github.com/nfnt/resize"
	"image"
	"image/jpeg"
	"os"
	"testing"
)

func TestConvertToGray(t *testing.T) {
	img := loadImage(t, "../../../../files/test/lenna.jpg")

	imgGray := ConvertToGray(img)
	saveImage(t, imgGray)
}

func TestConvertToGrayArray(t *testing.T) {
	img := loadImage(t, "../../../../files/test/lenna-gray.jpg")
	resized := resize.Resize(9, 8, img, resize.Bilinear)
	pixels := ConvertToGrayArray(resized)

	for i := 0; i < len(pixels); i++ {
		for j := 0; j < len(pixels[i])-1; j++ {
			fmt.Printf("%v - ", pixels[i][j])
		}
	}
}

func loadImage(t *testing.T, uri string) image.Image {
	f, err := os.Open(uri)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	img, err := jpeg.Decode(f)
	if err != nil {
		t.Fatal(err)
	}
	return img
}

func saveImage(t *testing.T, img image.Image) {
	t.Helper()

	nf, err := os.Create("../../../../files/test/lenna-gray.jpg")
	if err != nil {
		t.Fatal(err)
	}
	defer nf.Close()

	if err = jpeg.Encode(nf, img, nil); err != nil {
		t.Fatal(err)
	}
}
