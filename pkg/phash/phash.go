package phash

import (
	"fmt"
	"github.com/corona10/goimagehash"
	"image"
	"math/bits"
)

func AverageHash(img image.Image) (uint64, error) {
	h, err := goimagehash.AverageHash(img)
	if err != nil {
		return 0, err
	}
	return h.GetHash(), nil
}

func DifferenceHash(img image.Image) (uint64, error) {
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

func Distance(lHash, rHash uint64) (int, error) {
	hamming := lHash ^ rHash
	return bits.OnesCount64(hamming), nil
}
