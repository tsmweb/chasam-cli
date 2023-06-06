package media

import (
	"errors"
	"fmt"
	"image"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/tsmweb/chasam/app/hash"
	"github.com/tsmweb/chasam/common/mediautil"
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
	aHash       uint64
	mHash       uint64
	dHash       uint64
	dHashV      uint64
	dHashD      uint64
	pHash       uint64
	lHash       uint64
	wHash       uint64
	match       []Match
}

// NewMedia creates and returns a new Media instance.
func NewMedia(path string, hashTypes []hash.Type) (*Media, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("Media::NewMedia(%s) | Error: %v", path, err)
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
		return nil, fmt.Errorf("Media::NewMedia(%s) | Error: %v", path, err)
	}

	_, name := filepath.Split(info.Name())

	m := new(Media)
	m.path = path
	m.name = name
	m.modifiedAt = info.ModTime()
	m.mediaType = strings.Split(contentType.String(), "/")[0]
	m.contentType = contentType.String()

	var img image.Image
	getImg := func() image.Image {
		if img == nil {
			img, err = mediautil.Decode(file, mediautil.ContentType(m.contentType))
			if err != nil {
				return nil
			}
		}
		return img
	}

	for _, h := range hashTypes {
		switch h {
		case hash.SHA1:
			if err = m.setSHA1(file); err != nil {
				return nil, err
			}
		case hash.ED2K:
			if err = m.setED2K(file); err != nil {
				return nil, err
			}
		case hash.AHash:
			if err = m.setAHash(getImg()); err != nil {
				return nil, err
			}
		case hash.MHash:
			if err = m.setMHash(getImg()); err != nil {
				return nil, err
			}
		case hash.DHash:
			if err = m.setDHash(getImg()); err != nil {
				return nil, err
			}
		case hash.DHashV:
			if err = m.setDHashV(getImg()); err != nil {
				return nil, err
			}
		case hash.DHashD:
			if err = m.setDHashD(getImg()); err != nil {
				return nil, err
			}
		case hash.PHash:
			if err = m.setPHash(getImg()); err != nil {
				return nil, err
			}
		case hash.LHash:
			if err = m.setLHash(getImg()); err != nil {
				return nil, err
			}
		default:
			return nil, errors.New("hash not found")
		}
	}

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

func (m *Media) SHA1() string {
	return m.sha1
}

func (m *Media) ED2K() string {
	return m.ed2k
}

func (m *Media) AHash() uint64 {
	return m.aHash
}

func (m *Media) MHash() uint64 {
	return m.mHash
}

func (m *Media) DHash() uint64 {
	return m.dHash
}

func (m *Media) DHashV() uint64 {
	return m.dHashV
}

func (m *Media) DHashD() uint64 {
	return m.dHashD
}

func (m *Media) PHash() uint64 {
	return m.pHash
}

func (m *Media) LHash() uint64 {
	return m.lHash
}

func (m *Media) WHash() uint64 {
	return 0
}

func (m *Media) setSHA1(f *os.File) error {
	h, err := hash.Sha1Hash(f)
	if err != nil {
		return fmt.Errorf("Media::setSHA1(%s) | Error: %v", m.path, err)
	}
	m.sha1 = h
	return nil
}

func (m *Media) setED2K(f *os.File) error {
	h, err := hash.Ed2kHash(f)
	if err != nil {
		return fmt.Errorf("Media::setED2K(%s) | Error: %v", m.path, err)
	}
	m.ed2k = h
	return nil
}

func (m *Media) setAHash(img image.Image) error {
	h, err := hash.AverageHash(img)
	if err != nil {
		return fmt.Errorf("Media::setAHash(%s) | Error: %v", m.path, err)
	}
	m.aHash = h
	return nil
}

func (m *Media) setMHash(img image.Image) error {
	h, err := hash.ModeHash(img)
	if err != nil {
		return fmt.Errorf("Media::setMHash(%s) | Error: %v", m.path, err)
	}
	m.mHash = h
	return nil
}

func (m *Media) setDHash(img image.Image) error {
	h, err := hash.DifferenceHash(img)
	if err != nil {
		return fmt.Errorf("Media::setDHash(%s) | Error: %v", m.path, err)
	}
	m.dHash = h
	return nil
}

func (m *Media) setDHashV(img image.Image) error {
	h, err := hash.DifferenceHashVertical(img)
	if err != nil {
		return fmt.Errorf("Media::setDHashV(%s) | Error: %v", m.path, err)
	}
	m.dHashV = h
	return nil
}

func (m *Media) setDHashD(img image.Image) error {
	h, err := hash.DifferenceHashDiagonal(img)
	if err != nil {
		return fmt.Errorf("Media::setDHashD(%s) | Error: %v", m.path, err)
	}
	m.dHashD = h
	return nil
}

func (m *Media) setPHash(img image.Image) error {
	h, err := hash.PerceptionHash(img)
	if err != nil {
		return fmt.Errorf("Media::setPHash(%s) | Error: %v", m.path, err)
	}
	m.pHash = h
	return nil
}

func (m *Media) setLHash(img image.Image) error {
	h, err := hash.LeonardHash(img)
	if err != nil {
		return fmt.Errorf("Media::setLHash(%s) | Error: %v", m.path, err)
	}
	m.lHash = h
	return nil
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
