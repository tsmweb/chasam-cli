package phash

import (
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/corona10/goimagehash"
	"github.com/nfnt/resize"
	"github.com/tsmweb/chasam/pkg/phash/imgutil"
	"image"
	"math/bits"
)

// AverageHash function returns a hash computation of average hash vertically.
// Implementation follows
// https://www.hackerfactor.com/blog/index.php?/archives/432-Looks-Like-It.html
func AverageHash(img image.Image) (uint64, error) {
	if img == nil {
		return 0, errors.New("image cannot be nil")
	}

	w, h := 8, 8
	resized := resize.Resize(uint(w), uint(h), img, resize.Bilinear)
	pixels := imgutil.ConvertToGrayArray(resized)
	flatPixels := make([]float64, w*h)
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

// DifferenceHash function returns a hash computation of difference hash.
// Implementation follows
// https://www.hackerfactor.com/blog/index.php?/archives/529-Kind-of-Like-That.html
func DifferenceHash(img image.Image) (uint64, error) {
	if img == nil {
		return 0, errors.New("image cannot be nil")
	}

	w, h := 9, 8
	resized := resize.Resize(uint(w), uint(h), img, resize.Bilinear) // testar resize.Bicubic
	pixels := imgutil.ConvertToGrayArray(resized)
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
	pixels := imgutil.ConvertToGrayArray(resized)
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

// PerceptionHash function returns a hash computation of perception hash vertically.
// Implementation follows
// https://www.hackerfactor.com/blog/index.php?/archives/432-Looks-Like-It.html
func PerceptionHash(img image.Image) (uint64, error) {
	if img == nil {
		return 0, errors.New("image cannot be nil")
	}

	w, h := 32, 32
	resized := resize.Resize(uint(w), uint(h), img, resize.Bilinear)
	pixels := imgutil.ConvertToGrayArray(resized)
	dct := imgutil.DCT2D(pixels, w, h)

	// calculate the average of the dct.
	w, h = 8, 8
	flatDct := make([]float64, w*h)
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

func PerceptionHash_(img image.Image) (uint64, error) {
	h, err := goimagehash.PerceptionHash(img)
	if err != nil {
		return 0, err
	}
	return h.GetHash(), err
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
