package pathutil

import (
	"github.com/tsmweb/chasam/common/mediautil"
	"io/fs"
	"os"
	"path/filepath"
)

func GetTotalFiles(dir string) (int64, error) {
	var total int64
	err := filepath.Walk(dir, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		// checks if it is valid media.
		_, err = mediautil.GetContentType(file)
		if err != nil {
			return nil
		}

		total++

		return nil
	})
	if err != nil {
		return 0, err
	}

	return total, nil
}
