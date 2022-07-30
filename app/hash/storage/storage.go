package storage

import (
	"errors"
	"github.com/tsmweb/chasam/app/hash"
	"github.com/tsmweb/chasam/app/media"
	"github.com/tsmweb/chasam/pkg/phash"
	"os"
	"path/filepath"
)

type Storage struct {
	hashTable  map[hash.Type]map[string]string
	pHashTable map[hash.Type]map[uint64]string
}

func (s *Storage) AppendHash(hashType hash.Type, hashValue string, fileName string) {
	hashMedia, ok := s.hashTable[hashType]
	if !ok {
		hashMedia = make(map[string]string)
		s.hashTable[hashType] = hashMedia
	}
	hashMedia[hashValue] = fileName
}

func (s *Storage) FindByHash(hashType hash.Type, hashValue string) string {
	v, ok := s.hashTable[hashType][hashValue]
	if ok {
		return v
	}
	return "-1"
}

func (s *Storage) AppendPerceptualHash(hashType hash.Type, hashValue uint64, fileName string) {
	hashMedia, ok := s.pHashTable[hashType]
	if !ok {
		hashMedia = make(map[uint64]string)
		s.pHashTable[hashType] = hashMedia
	}
	hashMedia[hashValue] = fileName
}

func (s *Storage) FindByPerceptualHash(hashType hash.Type, hashValue uint64, distance int) (int, string) {
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

func NewStorage(dir string, hashTypes []hash.Type) (*Storage, error) {
	f, err := os.Open(dir)
	if err != nil {
		return nil, err
	}

	entries, _ := f.Readdir(-1)
	if len(entries) <= 0 {
		return nil, errors.New("images/videos not found")
	}

	storage := &Storage{
		hashTable:  make(map[hash.Type]map[string]string),
		pHashTable: make(map[hash.Type]map[uint64]string),
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		path := filepath.Join(dir, entry.Name())
		m, err := media.New(path, hashTypes)
		if err != nil {
			return nil, err
		}

		for _, typeHash := range hashTypes {
			switch typeHash {
			case hash.SHA1:
				if h := m.SHA1(); h != "" {
					storage.AppendHash(hash.SHA1, h, m.Name())
				}
			case hash.ED2K:
				if h := m.ED2K(); h != "" {
					storage.AppendHash(hash.ED2K, h, m.Name())
				}
			case hash.AHash:
				if h := m.AHash(); h > 0 {
					storage.AppendPerceptualHash(hash.AHash, h, m.Name())
				}
			case hash.DHash:
				if h := m.DHash(); h > 0 {
					storage.AppendPerceptualHash(hash.DHash, h, m.Name())
				}
			case hash.DHashV:
				if h := m.DHashV(); h > 0 {
					storage.AppendPerceptualHash(hash.DHashV, h, m.Name())
				}
			case hash.PHash:
				if h := m.PHash(); h > 0 {
					storage.AppendPerceptualHash(hash.PHash, h, m.Name())
				}
			case hash.WHash:
				return nil, errors.New("w-hash not implemented")
			default:
				return nil, errors.New("invalid hash")
			}
		}
	}

	return storage, nil
}
