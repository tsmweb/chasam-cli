package hash

import (
	"errors"
	"github.com/tsmweb/chasam/app/media"
	"github.com/tsmweb/chasam/pkg/phash"
	"os"
	"path/filepath"
)

type Type int

const (
	SHA1 Type = iota
	ED2K
	AHash
	DHash
	DHashV
	PHash
	WHash
)

type Storage struct {
	hashTable  map[Type]map[string]string
	pHashTable map[Type]map[uint64]string
}

func (s *Storage) AppendHash(hashType Type, hashValue string, fileName string) {
	hashMedia, ok := s.hashTable[hashType]
	if !ok {
		hashMedia = make(map[string]string)
		s.hashTable[hashType] = hashMedia
	}
	hashMedia[hashValue] = fileName
}

func (s *Storage) FindByHash(hashType Type, hashValue string) string {
	v, ok := s.hashTable[hashType][hashValue]
	if ok {
		return v
	}
	return "-1"
}

func (s *Storage) AppendPerceptualHash(hashType Type, hashValue uint64, fileName string) {
	hashMedia, ok := s.pHashTable[hashType]
	if !ok {
		hashMedia = make(map[uint64]string)
		s.pHashTable[hashType] = hashMedia
	}
	hashMedia[hashValue] = fileName
}

func (s *Storage) FindByPerceptualHash(hashType Type, hashValue uint64, distance int) (int, string) {
	for lHash, fName := range s.pHashTable[hashType] {
		dist, err := phash.Distance(lHash, hashValue)
		if err != nil {
			return -1, ""
		}

		if dist <= distance {
			return dist, fName
		}
	}

	return -1, ""
}

func NewStorage(dir string, hashTypes []Type) (*Storage, error) {
	f, err := os.Open(dir)
	if err != nil {
		return nil, err
	}

	entries, _ := f.Readdir(-1)
	if len(entries) <= 0 {
		return nil, errors.New("images/videos not found")
	}

	storage := &Storage{
		hashTable:  make(map[Type]map[string]string),
		pHashTable: make(map[Type]map[uint64]string),
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		path := filepath.Join(dir, entry.Name())
		m, err := media.New(path)
		if err != nil {
			return nil, err
		}

		for _, typeHash := range hashTypes {
			switch typeHash {
			case SHA1:
				if h, err := m.SHA1(); err == nil {
					storage.AppendHash(SHA1, h, m.Name())
				}
			case ED2K:
				if h, err := m.ED2K(); err == nil {
					storage.AppendHash(ED2K, h, m.Name())
				}
			case AHash:
				if h, err := m.AHash(); err == nil {
					storage.AppendPerceptualHash(AHash, h[0], m.Name())
				}
			case DHash:
				if h, err := m.DHash(); err == nil {
					storage.AppendPerceptualHash(DHash, h[0], m.Name())
				}
			case DHashV:
				if h, err := m.DHashV(); err == nil {
					storage.AppendPerceptualHash(DHashV, h[0], m.Name())
				}
			case PHash:
				if h, err := m.PHash(); err == nil {
					storage.AppendPerceptualHash(PHash, h[0], m.Name())
				}
			case WHash:
				return nil, errors.New("w-hash not implemented")
			default:
				return nil, errors.New("invalid hash")
			}
		}
	}

	return storage, nil
}
