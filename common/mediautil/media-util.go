package mediautil

import (
	"errors"
	"fmt"
	"github.com/nfnt/resize"
	"golang.org/x/image/bmp"
	"golang.org/x/image/tiff"
	"golang.org/x/image/webp"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"os"
)

type ContentType string

func (ct ContentType) String() string {
	return string(ct)
}

const (
	ImageGIF  ContentType = "image/gif"
	ImageJPEG ContentType = "image/jpeg"
	ImagePNG  ContentType = "image/png"
	ImageBMP  ContentType = "image/bmp"
	ImageWEBP ContentType = "image/webp"
	ImageTIFF ContentType = "image/tiff"

	VideoMPEG ContentType = "video/mpeg"
	VideoMP4  ContentType = "video/mp4"
	VideoAVI  ContentType = "video/avi"
	VideoOGG  ContentType = "video/ogg"
	VideoWEBM ContentType = "video/webm"
	VideoWMV  ContentType = "video/wmv"
	VideoFLV  ContentType = "video/flv"
	VideoMKV  ContentType = "video/mkv"
	VideoMOV  ContentType = "video/mov"
)

var ErrUnsupportedMediaType = errors.New("unsupported media type")

func GetContentType(out *os.File) (contentType ContentType, err error) {
	fileHeader := make([]byte, 512)

	if _, err = out.Read(fileHeader); err != nil {
		return
	}
	if _, err = out.Seek(0, io.SeekStart); err != nil {
		return
	}

	_contentType := DetectContentType(fileHeader)

	switch _contentType {
	case ImageGIF.String():
		contentType = ImageGIF

	case ImageJPEG.String():
		contentType = ImageJPEG

	case ImagePNG.String():
		contentType = ImagePNG

	case ImageBMP.String():
		contentType = ImageBMP

	case ImageWEBP.String():
		contentType = ImageWEBP

	case ImageTIFF.String():
		contentType = ImageTIFF

	case VideoMPEG.String():
		contentType = VideoMPEG

	case VideoMP4.String():
		contentType = VideoMP4

	case VideoAVI.String():
		contentType = VideoAVI

	case VideoOGG.String():
		contentType = VideoOGG

	case VideoWEBM.String():
		contentType = VideoWEBM

	case VideoWMV.String():
		contentType = VideoWMV

	case VideoFLV.String():
		contentType = VideoFLV

	case VideoMKV.String():
		contentType = VideoMKV

	case VideoMOV.String():
		contentType = VideoMOV

	default:
		err = ErrUnsupportedMediaType
	}

	return
}

func Encode(w io.Writer, m image.Image, t string) (err error) {
	switch t {
	case ImageGIF.String():
		err = gif.Encode(w, m, nil)
	case ImageJPEG.String():
		err = jpeg.Encode(w, m, &jpeg.Options{Quality: jpeg.DefaultQuality})
	case ImagePNG.String():
		err = png.Encode(w, m)
	case ImageBMP.String():
		err = bmp.Encode(w, m)
	case ImageTIFF.String():
		err = tiff.Encode(w, m, nil)
	default:
		err = ErrUnsupportedMediaType
	}

	return
}

func Decode(f *os.File, t string) (img image.Image, err error) {
	switch t {
	case ImageGIF.String():
		img, err = gif.Decode(f)
	case ImageJPEG.String():
		img, err = jpeg.Decode(f)
	case ImagePNG.String():
		img, err = png.Decode(f)
	case ImageBMP.String():
		img, err = bmp.Decode(f)
	case ImageWEBP.String():
		img, err = webp.Decode(f)
	case ImageTIFF.String():
		img, err = tiff.Decode(f)
	default:
		err = fmt.Errorf("unrecognized file")
	}

	return
}

func Resize(img image.Image, with, height uint) image.Image {
	return resize.Resize(with, height, img, resize.Lanczos3)
}
