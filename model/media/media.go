package media

import (
	"chasam/utils/hashutil"
	"chasam/utils/imageutil"
	"os"
	"time"
)

// MediaType represents the type of media, such as Image or Video.
type MediaType int

const (
	MediaTypeImage MediaType = 0x1
	MediaTypeVideo MediaType = 0x2
)

func (mt MediaType) String() (str string) {
	name := func(mediaType MediaType, name string) bool {
		if mt&mediaType == 0 {
			return false
		}
		str = name
		return true
	}

	if name(MediaTypeImage, "image") {
		return
	}
	if name(MediaTypeVideo, "video") {
		return
	}
	return
}

// Media represents the information of a media and its hash.
type Media struct {
	Name        string
	Path        string
	Type        string
	KeywordList []string
	HashED2K    string
	HashSHA1    string
	HashA       string
	HashD       string
	HashP       string
	HashW       string
	ModifiedAt  time.Time
}

// New creates and returns a new Media instance.
func New(file *os.File) (*Media, error) {
	info, err := file.Stat()
	if err != nil {
		return nil, err
	}

	mediaType, err := imageutil.GetMediaType(file)
	if err != nil {
		return nil, err
	}

	hED2K, err := hashutil.HashED2K(file)
	if err != nil {
		return nil, err
	}

	hSha1, err := hashutil.HashSHA1(file)
	if err != nil {
		return nil, err
	}

	media := new(Media)
	media.Name = info.Name()
	media.Path = file.Name()
	media.Type = mediaType.String()
	media.HashED2K = hED2K
	media.HashSHA1 = hSha1
	media.ModifiedAt = info.ModTime()

	return media, nil
}
