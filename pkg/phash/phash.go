package phash

import (
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/corona10/goimagehash"
	"github.com/nfnt/resize"
	"github.com/tsmweb/chasam/pkg/phash/imageutil"
	"image"
	"math/bits"
)

// DifferenceHash function returns a hash computation of difference hash.
// Implementation follows
// https://www.hackerfactor.com/blog/index.php?/archives/529-Kind-of-Like-That.html
func DifferenceHash(img image.Image) (uint64, error) {
	if img == nil {
		return 0, errors.New("image cannot be nil")
	}

	resized := resize.Resize(9, 8, img, resize.Bilinear) // testar resize.Bicubic
	pixels := imageutil.ConvertToGrayArray(resized)
	idx := 0
	var hash uint64

	for y := 0; y < len(pixels); y++ {
		for x := 0; x < len(pixels[y])-1; x++ {
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

	resized := resize.Resize(8, 9, img, resize.Bilinear)
	pixels := imageutil.ConvertToGrayArray(resized)
	idx := 0
	var hash uint64

	for y := 0; y < len(pixels)-1; y++ {
		for x := 0; x < len(pixels[y]); x++ {
			if pixels[y][x] < pixels[y+1][x] {
				hash |= 1 << uint(64-idx-1)
			}
			idx++
		}
	}

	return hash, nil
}

func AverageHash(img image.Image) (uint64, error) {
	h, err := goimagehash.AverageHash(img)
	if err != nil {
		return 0, err
	}
	return h.GetHash(), nil
}

func DifferenceHash_(img image.Image) (uint64, error) {
	h, err := goimagehash.DifferenceHash(img)
	if err != nil {
		return 0, err
	}
	return h.GetHash(), nil
}

func PerceptionHash(img image.Image) (uint64, error) {
	h, err := goimagehash.PerceptionHash(img)
	if err != nil {
		return 0, err
	}
	return h.GetHash(), err
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
