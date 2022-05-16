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

type MediaType string

func (mt MediaType) String() string {
	return string(mt)
}

const (
	ImageGIF  MediaType = "image/gif"  // implement
	ImageJPEG MediaType = "image/jpeg" // implement
	ImagePNG  MediaType = "image/png"  // implement
	ImageBMP  MediaType = "image/bmp"  // implement
	ImageWEBP MediaType = "image/webp" // implement
	ImageTIFF MediaType = "image/tiff" // implement

	VideoMPEG MediaType = "video/mpeg"      // implement
	VideoMP4  MediaType = "video/mp4"       // implement
	VideoAVI  MediaType = "video/avi"       // implement
	VideoOGG  MediaType = "application/ogg" // implement
	VideoWEBM MediaType = "video/webm"      // implement
	VideoWMV  MediaType = "video/wmv"       // implement
	VideoFLV  MediaType = "video/flv"       // implement
	VideoMKV  MediaType = "video/mkv"       // implement
	VideoMOV  MediaType = "video/mov"       //implement
)

var ErrUnsupportedMediaType = errors.New("unsupported media type")

func GetMediaType(out *os.File) (mediaType MediaType, err error) {
	fileHeader := make([]byte, 512)

	if _, err = out.Read(fileHeader); err != nil {
		return
	}
	if _, err = out.Seek(0, io.SeekStart); err != nil {
		return
	}

	contentType := DetectContentType(fileHeader)

	switch contentType {
	case ImageGIF.String():
		mediaType = ImageGIF

	case ImageJPEG.String():
		mediaType = ImageJPEG

	case ImagePNG.String():
		mediaType = ImagePNG

	case ImageBMP.String():
		mediaType = ImageBMP

	case ImageWEBP.String():
		mediaType = ImageWEBP

	case ImageTIFF.String():
		mediaType = ImageTIFF

	case VideoMPEG.String():
		mediaType = VideoMPEG

	case VideoMP4.String():
		mediaType = VideoMP4

	case VideoAVI.String():
		mediaType = VideoAVI

	case VideoOGG.String():
		mediaType = VideoOGG

	case VideoWEBM.String():
		mediaType = VideoWEBM

	case VideoWMV.String():
		mediaType = VideoWMV

	case VideoFLV.String():
		mediaType = VideoFLV

	case VideoMKV.String():
		mediaType = VideoMKV

	case VideoMOV.String():
		mediaType = VideoMOV

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
