package media

import (
	"fmt"
	"github.com/tsmweb/chasam/common/hashutil"
	"github.com/tsmweb/chasam/common/mediautil"
	"github.com/tsmweb/chasam/pkg/phash"
	"image"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Match struct {
	Name     string
	HashType string
	Distance int
}

// Media represents the information of a media and its hash.
type Media struct {
	name        string
	path        string
	mediaType   string
	contentType string
	modifiedAt  time.Time
	ed2k        string
	sha1        string
	aHash       []uint64
	dHash       []uint64
	dHashV      []uint64
	pHash       []uint64
	wHash       []uint64
	match       []Match
}

// New creates and returns a new Media instance.
func New(path string) (*Media, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("Media::New(%s) | Error: %v", path, err)
	}
	defer file.Close()

	// checks if it is valid media.
	contentType, err := mediautil.GetContentType(file)
	if err != nil {
		return nil, err
	}

	// get file information.
	info, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("Media::New(%s) | Error: %v", path, err)
	}

	_, name := filepath.Split(info.Name())

	m := new(Media)
	m.path = path
	m.name = name
	m.modifiedAt = info.ModTime()
	m.mediaType = strings.Split(contentType.String(), "/")[0]
	m.contentType = contentType.String()

	return m, nil
}

func (m *Media) Name() string {
	return m.name
}

func (m *Media) Path() string {
	return m.path
}

func (m *Media) Type() string {
	return m.mediaType
}

func (m *Media) ContentType() string {
	return m.contentType
}

func (m *Media) ModifiedAt() time.Time {
	return m.modifiedAt
}

func (m *Media) SHA1() (string, error) {
	if m.sha1 != "" {
		return m.sha1, nil
	}

	f, err := os.Open(m.path)
	if err != nil {
		return "", fmt.Errorf("Media::SHA1(%s) | Error: %v", m.path, err)
	}
	defer f.Close()

	h, err := hashutil.HashSHA1(f)
	if err != nil {
		return "", fmt.Errorf("Media::SHA1(%s) | Error: %v", m.path, err)
	}
	m.sha1 = h

	return h, nil
}

func (m *Media) ED2K() (string, error) {
	if m.ed2k != "" {
		return m.ed2k, nil
	}

	f, err := os.Open(m.path)
	if err != nil {
		return "", fmt.Errorf("Media::ED2K(%s) | Error: %v", m.path, err)
	}
	defer f.Close()

	h, err := hashutil.HashED2K(f)
	if err != nil {
		return "", fmt.Errorf("Media::ED2K(%s) | Error: %v", m.path, err)
	}
	m.ed2k = h

	return h, nil
}

func (m *Media) AHash() ([]uint64, error) {
	if m.aHash != nil {
		return m.aHash, nil
	}

	img, err := m.getImage()
	if err != nil {
		return nil, fmt.Errorf("Media::AHash(%s) | Error: %v", m.path, err)
	}

	h, err := phash.AverageHash(img)
	if err != nil {
		return nil, fmt.Errorf("Media::AHash(%s) | Error: %v", m.path, err)
	}
	ah := []uint64{h}
	m.aHash = ah

	return ah, nil
}

func (m *Media) DHash() ([]uint64, error) {
	if m.dHash != nil {
		return m.dHash, nil
	}

	img, err := m.getImage()
	if err != nil {
		return nil, fmt.Errorf("Media::DHash(%s) | Error: %v", m.path, err)
	}

	h, err := phash.DifferenceHash(img)
	if err != nil {
		return nil, fmt.Errorf("Media::DHash(%s) | Error: %v", m.path, err)
	}
	dh := []uint64{h}
	m.dHash = dh

	return dh, nil
}

func (m *Media) DHashV() ([]uint64, error) {
	if m.dHashV != nil {
		return m.dHashV, nil
	}

	img, err := m.getImage()
	if err != nil {
		return nil, fmt.Errorf("Media::DHashV(%s) | Error: %v", m.path, err)
	}

	h, err := phash.DifferenceHashVertical(img)
	if err != nil {
		return nil, fmt.Errorf("Media::DHashV(%s) | Error: %v", m.path, err)
	}
	dhv := []uint64{h}
	m.dHashV = dhv

	return dhv, nil
}

func (m *Media) PHash() ([]uint64, error) {
	if m.pHash != nil {
		return m.pHash, nil
	}

	img, err := m.getImage()
	if err != nil {
		return nil, fmt.Errorf("Media::PHash(%s) | Error: %v", m.path, err)
	}

	h, err := phash.PerceptionHash(img)
	if err != nil {
		return nil, fmt.Errorf("Media::PHash(%s) | Error: %v", m.path, err)
	}
	ph := []uint64{h}
	m.pHash = ph

	return ph, nil
}

func (m *Media) WHash() ([]uint64, error) {
	if m.wHash != nil {
		return m.wHash, nil
	}

	wh := []uint64{0}
	m.wHash = wh

	return wh, nil
}

func (m *Media) AddMatch(name string, hashType string, distance int) {
	m.match = append(m.match, Match{
		Name:     name,
		HashType: hashType,
		Distance: distance,
	})
}

func (m *Media) Match() []Match {
	return m.match
}

func (m *Media) getImage() (image.Image, error) {
	f, err := os.Open(m.path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	img, err := mediautil.Decode(f, mediautil.ContentType(m.contentType))
	if err != nil {
		return nil, err
	}

	return img, nil
}
