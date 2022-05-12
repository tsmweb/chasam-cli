package hashutil

import (
	"bufio"
	"chasam/utils/hashutil/ed2k"
	"crypto/sha1"
	"fmt"
	"io"
)

func HashSHA1(r io.Reader) (string, error) {
	rd := bufio.NewReader(r)
	sh := sha1.New()
	_, err := rd.WriteTo(sh)
	if err != nil {
		return "", err
	}
	h := fmt.Sprintf("%x", sh.Sum(nil))
	return h, nil
}

func HashED2K(r io.Reader) (string, error) {
	rd := bufio.NewReader(r)
	sh := ed2k.New()
	_, err := rd.WriteTo(sh)
	if err != nil {
		return "", err
	}
	h := fmt.Sprintf("%x", sh.Sum(nil))
	return h, nil
}
