package hashutil

import (
	"bufio"
	"crypto/sha1"
	"fmt"
	"github.com/tsmweb/chasam/pkg/ed2k"
	"io"
	"os"
)

func HashSHA1(f *os.File) (string, error) {
	rd := bufio.NewReader(f)
	sh := sha1.New()
	_, err := rd.WriteTo(sh)
	if err != nil {
		return "", err
	}
	h := fmt.Sprintf("%x", sh.Sum(nil))

	if _, err = f.Seek(0, io.SeekStart); err != nil {
		return h, err
	}

	return h, nil
}

func HashED2K(f *os.File) (string, error) {
	rd := bufio.NewReader(f)
	sh := ed2k.New()
	_, err := rd.WriteTo(sh)
	if err != nil {
		return "", err
	}
	h := fmt.Sprintf("%x", sh.Sum(nil))

	if _, err = f.Seek(0, io.SeekStart); err != nil {
		return h, err
	}

	return h, nil
}
