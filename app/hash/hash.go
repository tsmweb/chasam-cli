package hash

import (
	"bufio"
	"crypto/sha1"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"image"
	"io"
	"math/bits"
	"os"

	"github.com/nfnt/resize"
	"github.com/tsmweb/chasam/app/hash/transform"
	"github.com/tsmweb/chasam/pkg/ed2k"
)

type Type int

const (
	SHA1 Type = iota
	ED2K
	AHash
	MHash
	DHash
	DHashV
	DHashD
	PHash
	LHash
	WHash
)

func (t Type) String() string {
	switch t {
	case SHA1:
		return "SHA1"
	case ED2K:
		return "ED2K"
	case AHash:
		return "AHash"
	case MHash:
		return "MHash"
	case DHash:
		return "DHash"
	case DHashV:
		return "DHashV"
	case DHashD:
		return "DHashD"
	case PHash:
		return "PHash"
	case LHash:
		return "LHash"
	case WHash:
		return "WHash"
	default:
		return ""
	}
}

func Sha1Hash(f *os.File) (string, error) {
	if err := seekStart(f); err != nil {
		return "", err
	}

	rd := bufio.NewReader(f)
	sh := sha1.New()
	_, err := rd.WriteTo(sh)
	if err != nil {
		return "", err
	}
	h := fmt.Sprintf("%x", sh.Sum(nil))

	if err = seekStart(f); err != nil {
		return "", err
	}

	return h, nil
}

func Ed2kHash(f *os.File) (string, error) {
	if err := seekStart(f); err != nil {
		return "", err
	}

	rd := bufio.NewReader(f)
	sh := ed2k.New()
	_, err := rd.WriteTo(sh)
	if err != nil {
		return "", err
	}
	h := fmt.Sprintf("%x", sh.Sum(nil))

	if err = seekStart(f); err != nil {
		return "", err
	}

	return h, nil
}

func seekStart(f *os.File) error {
	_, err := f.Seek(0, io.SeekStart)
	return err
}

// AverageHash function returns a hash computation of average hash vertically.
// Implementation follows
// https://www.hackerfactor.com/blog/index.php?/archives/432-Looks-Like-It.html
func AverageHash(img image.Image) (uint64, error) {
	if img == nil {
		return 0, errors.New("image cannot be nil")
	}

	w, h := 8, 8
	resized := resize.Resize(uint(w), uint(h), img, resize.Bilinear)
	pixels := transform.ConvertToGrayArray(resized)
	flatPixels := [64]float64{}
	sum := 0.0

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			sum += pixels[y][x]
			flatPixels[h*y+x] = pixels[y][x]
		}
	}

	avg := sum / float64(64)
	var hash uint64

	for idx, p := range flatPixels {
		if p > avg {
			hash |= 1 << uint(64-idx-1)
		}
	}

	return hash, nil
}

func ModeHash(img image.Image) (uint64, error) {
	if img == nil {
		return 0, errors.New("image cannot be nil")
	}

	w, h := 8, 8
	resized := resize.Resize(uint(w), uint(h), img, resize.Bilinear)
	pixels := transform.ConvertToGrayArray(resized)
	flatPixels := [64]float64{}

	countMap := make(map[int]int)

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			pixel := pixels[y][x]
			countMap[int(pixel)]++
			flatPixels[h*y+x] = pixel
		}
	}

	var pixel int
	max := 0

	for p, c := range countMap {
		if c > max {
			max = c
			pixel = p
		}
	}

	var hash uint64

	for idx, p := range flatPixels {
		if int(p) < pixel {
			hash |= 1 << uint(64-idx-1)
		}
	}

	return hash, nil
}

// DifferenceHash function returns a hash computation of difference hash.
// Implementation follows
// https://www.hackerfactor.com/blog/index.php?/archives/529-Kind-of-Like-That.html
func DifferenceHash(img image.Image) (uint64, error) {
	if img == nil {
		return 0, errors.New("image cannot be nil")
	}

	w, h := 9, 8
	resized := resize.Resize(uint(w), uint(h), img, resize.Bilinear) // testar resize.Bicubic
	pixels := transform.ConvertToGrayArray(resized)
	idx := 0
	var hash uint64

	for y := 0; y < h; y++ {
		for x := 0; x < w-1; x++ {
			if pixels[y][x] < pixels[y][x+1] {
				hash |= 1 << uint(64-idx-1)
			}
			idx++
		}
	}

	return hash, nil
}

// DifferenceHashVertical function returns a hash computation of difference hash vertically.
// Implementation follows
// https://www.hackerfactor.com/blog/index.php?/archives/529-Kind-of-Like-That.html
func DifferenceHashVertical(img image.Image) (uint64, error) {
	if img == nil {
		return 0, errors.New("image cannot be nil")
	}

	w, h := 8, 9
	resized := resize.Resize(uint(w), uint(h), img, resize.Bilinear)
	pixels := transform.ConvertToGrayArray(resized)
	idx := 0
	var hash uint64

	for y := 0; y < h-1; y++ {
		for x := 0; x < w; x++ {
			if pixels[y][x] < pixels[y+1][x] {
				hash |= 1 << uint(64-idx-1)
			}
			idx++
		}
	}

	return hash, nil
}

func DifferenceHashDiagonal(img image.Image) (uint64, error) {
	if img == nil {
		return 0, errors.New("image cannot be nil")
	}

	w, h := 9, 9
	resized := resize.Resize(uint(w), uint(h), img, resize.Bilinear) // testar resize.Bicubic
	pixels := transform.ConvertToGrayArray(resized)
	idx := 0
	var hash uint64

	for x := w - 1; x >= 0; x-- {
		for y := 0; y < (w - x - 1); y++ {
			_x := x + y
			if pixels[y][_x] > pixels[y+1][_x+1] {
				hash |= 1 << uint(64-idx-1)
			}
			idx++
		}
	}

	for y := h - 1; y > 0; y-- {
		for x := 0; x < (w - y - 1); x++ {
			_y := y + x
			if pixels[_y][x] > pixels[_y+1][x+1] {
				hash |= 1 << uint(64-idx-1)
			}
			idx++
		}
	}

	return hash, nil
}

// PerceptionHash function returns a hash computation of perception hash vertically.
// Implementation follows
// https://www.hackerfactor.com/blog/index.php?/archives/432-Looks-Like-It.html
func PerceptionHash(img image.Image) (uint64, error) {
	if img == nil {
		return 0, errors.New("image cannot be nil")
	}

	w, h := 32, 32
	resized := resize.Resize(uint(w), uint(h), img, resize.Bilinear)
	pixels := transform.ConvertToGrayArray(resized)
	dct := transform.DCT2D(pixels, w, h)

	// calculate the average of the dct.
	w, h = 8, 8
	flatDct := [64]float64{} // 8x8
	sum := 0.0

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			sum += dct[y][x]
			flatDct[h*y+x] = dct[y][x]
		}
	}

	// excluding the first term since the DC coefficient can be significantly different from the
	// other values and will throw off the average.
	sum -= dct[0][0]
	avg := sum / float64(63)

	// extract the hash.
	var hash uint64

	for idx, p := range flatDct {
		if p > avg {
			hash |= 1 << uint(64-idx-1)
		}
	}

	return hash, nil
}

func LeonardHash(img image.Image) (uint64, error) {
	if img == nil {
		return 0, errors.New("image cannot be nil")
	}

	w, h := 32, 32
	resized := resize.Resize(uint(w), uint(h), img, resize.Bilinear)
	pixels := transform.ConvertToThresholdArray(resized, 114)
	dct := transform.DCT2D(pixels, w, h)

	// calculate the average of the dct.
	w, h = 8, 8
	flatDct := [64]float64{} // 8x8
	sum := 0.0

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			sum += dct[y][x]
			flatDct[h*y+x] = dct[y][x]
		}
	}

	// excluding the first term since the DC coefficient can be significantly different from the
	// other values and will throw off the average.
	sum -= dct[0][0]
	avg := sum / float64(63)

	// extract the hash.
	var hash uint64

	for idx, p := range flatDct {
		if p > avg {
			hash |= 1 << uint(64-idx-1)
		}
	}

	return hash, nil
}

func WaveletHash(img image.Image) (uint64, error) {
	return 0, nil
}

func FormatToHex(hash uint64) string {
	return fmt.Sprintf("%016x", hash)
}

func ExtFormatToHex(hashs []uint64) string {
	var hexBytes []byte

	for _, hash := range hashs {
		hashBytes := make([]byte, 8)
		binary.BigEndian.PutUint64(hashBytes, hash)
		hexBytes = append(hexBytes, hashBytes...)
	}

	return hex.EncodeToString(hexBytes)
}

func Distance(lHash, rHash uint64) (int, error) {
	hamming := lHash ^ rHash
	return bits.OnesCount64(hamming), nil
}

func ExtDistance(lHash, rHash []uint64) (int, error) {
	if len(lHash) != len(rHash) {
		return -1, errors.New("extended image hash's size should be identical")
	}

	distance := 0

	for idx, lh := range lHash {
		rh := rHash[idx]
		hamming := lh ^ rh
		distance += bits.OnesCount64(hamming)
	}

	return distance, nil
}
