package media

import (
	"fmt"
	"github.com/tsmweb/chasam/common/imageutil"
	"os"
	"path/filepath"
	"time"
)

// Media represents the information of a media and its hash.
type Media struct {
	Name       string
	Path       string
	Type       string
	HashED2K   string
	HashSHA1   string
	HashA      string
	HashD      string
	HashP      string
	HashW      string
	ModifiedAt time.Time
}

// New creates and returns a new Media instance.
func New(path string) (*Media, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("Media::New(%s) | Error: %v", path, err)
	}
	defer file.Close()

	// checks if it is valid media.
	mediaType, err := imageutil.GetMediaType(file)
	if err != nil {
		return nil, err
	}

	// get file information.
	info, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("Media::New(%s) | Error: %v", path, err)
	}

	_, name := filepath.Split(info.Name())

	media := new(Media)
	media.Path = path
	media.Name = name
	media.ModifiedAt = info.ModTime()
	media.Type = mediaType.String()

	//hED2K, err := hashutil.HashED2K(file)
	//if err != nil {
	//	return nil, err
	//}
	//media.HashED2K = hED2K
	//
	//hSha1, err := hashutil.HashSHA1(file)
	//if err != nil {
	//	return nil, err
	//}
	//media.HashSHA1 = hSha1

	return media, nil
}
