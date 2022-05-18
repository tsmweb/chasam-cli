package media

import (
	"fmt"
	"github.com/tsmweb/chasam/common/hashutil"
	"github.com/tsmweb/chasam/common/mediautil"
	"github.com/tsmweb/chasam/pkg/phash"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Media represents the information of a media and its hash.
type Media struct {
	fd          *os.File
	Name        string
	Path        string
	Type        string
	ContentType string
	ModifiedAt  time.Time
	ED2K        string
	SHA1        string
	AHash       []uint64
	DHash       []uint64
	PHash       []uint64
	WHash       []uint64
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
	m.Path = path
	m.Name = name
	m.ModifiedAt = info.ModTime()
	m.Type = strings.Split(contentType.String(), "/")[0]
	m.ContentType = contentType.String()

	//----------------------------------------------------------------------------------------------
	if m.Type == "video" {
		return m, nil
	}

	sh, err := hashutil.HashSHA1(file)
	if err != nil {
		return nil, fmt.Errorf("Media::New(%s) | Error: %v", path, err)
	}
	m.SHA1 = sh

	eh, err := hashutil.HashED2K(file)
	if err != nil {
		return nil, fmt.Errorf("Media::New(%s) | Error: %v", path, err)
	}
	m.ED2K = eh

	img, err := mediautil.Decode(file, m.ContentType)
	if err != nil {
		return nil, fmt.Errorf("mediautil.Decode(%s) | Error: %v", m.Path, err)
	}

	ah, err := phash.AverageHash(img)
	if err != nil {
		return nil, fmt.Errorf("phash.AverageHash(%s) | Error: %v", m.Path, err)
	}
	m.AHash = []uint64{ah}

	dh, err := phash.DifferenceHash(img)
	if err != nil {
		return nil, fmt.Errorf("phash.DifferenceHash(%s) | Error: %v", m.Path, err)
	}
	m.DHash = []uint64{dh}

	ph, err := phash.PerceptionHash(img)
	if err != nil {
		return nil, fmt.Errorf("phash.PerceptionHash(%s) | Error: %v", m.Path, err)
	}
	m.PHash = []uint64{ph}
	//----------------------------------------------------------------------------------------------

	return m, nil
}

func (m *Media) Close() error {
	if m.fd != nil {
		return m.fd.Close()
	}
	return nil
}

func (m *Media) GenSHA1() error {
	f, err := os.Open(m.Path)
	if err != nil {
		return fmt.Errorf("Media::GenSHA1(%s) | Error: %v", m.Path, err)
	}
	defer f.Close()

	h, err := hashutil.HashSHA1(f)
	if err != nil {
		return fmt.Errorf("Media::GenSHA1(%s) | Error: %v", m.Path, err)
	}
	m.SHA1 = h

	return nil
}

func (m *Media) GenED2K() error {
	f, err := os.Open(m.Path)
	if err != nil {
		return fmt.Errorf("Media::GenED2K(%s) | Error: %v", m.Path, err)
	}
	defer f.Close()

	h, err := hashutil.HashED2K(f)
	if err != nil {
		return fmt.Errorf("Media::GenED2K(%s) | Error: %v", m.Path, err)
	}
	m.ED2K = h

	return nil
}

func (m *Media) GenAHash() error {
	if m.Type == "video" {
		return nil
	}

	f, err := os.Open(m.Path)
	if err != nil {
		return fmt.Errorf("Media::GenAHash(%s) | Error: %v", m.Path, err)
	}
	defer f.Close()

	img, err := mediautil.Decode(f, m.ContentType)
	if err != nil {
		return fmt.Errorf("Media::GenAHash(%s) | Error: %v", m.Path, err)
	}

	h, err := phash.AverageHash(img)
	if err != nil {
		return fmt.Errorf("Media::GenAHash(%s) | Error: %v", m.Path, err)
	}

	m.AHash = []uint64{h}
	return nil
}

func (m *Media) GenDHash() error {
	if m.Type == "video" {
		return nil
	}

	f, err := os.Open(m.Path)
	if err != nil {
		return fmt.Errorf("Media::GenDHash(%s) | Error: %v", m.Path, err)
	}
	defer f.Close()

	img, err := mediautil.Decode(f, m.ContentType)
	if err != nil {
		return fmt.Errorf("Media::GenDHash(%s) | Error: %v", m.Path, err)
	}

	h, err := phash.DifferenceHash(img)
	if err != nil {
		return fmt.Errorf("Media::GenDHash(%s) | Error: %v", m.Path, err)
	}

	m.DHash = []uint64{h}
	return nil
}

func (m *Media) GenPHash() error {
	if m.Type == "video" {
		return nil
	}

	f, err := os.Open(m.Path)
	if err != nil {
		return fmt.Errorf("Media::GenPHash(%s) | Error: %v", m.Path, err)
	}
	defer f.Close()

	img, err := mediautil.Decode(f, m.ContentType)
	if err != nil {
		return fmt.Errorf("Media::GenPHash(%s) | Error: %v", m.Path, err)
	}

	h, err := phash.PerceptionHash(img)
	if err != nil {
		return fmt.Errorf("Media::GenPHash(%s) | Error: %v", m.Path, err)
	}

	m.PHash = []uint64{h}
	return nil
}

func (m *Media) GenWHash() error {
	if m.Type == "video" {
		return nil
	}

	m.WHash = []uint64{0}
	return nil
}
