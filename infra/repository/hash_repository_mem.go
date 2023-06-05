package repository

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/tsmweb/chasam/app/hash"
	"github.com/tsmweb/chasam/app/media"
)

type mediaRepositoryMem struct {
	hashTable  map[hash.Type]map[string]string
	pHashTable map[hash.Type]map[uint64]string
}

func (r *mediaRepositoryMem) AppendHash(hashType hash.Type, hashValue string, fileName string) {
	hashMedia, ok := r.hashTable[hashType]
	if !ok {
		hashMedia = make(map[string]string)
		r.hashTable[hashType] = hashMedia
	}
	hashMedia[hashValue] = fileName
}

func (r *mediaRepositoryMem) FindByHash(hashType hash.Type, hashValue string) string {
	v, ok := r.hashTable[hashType][hashValue]
	if ok {
		return v
	}
	return "-1"
}

func (r *mediaRepositoryMem) AppendPerceptualHash(hashType hash.Type, hashValue uint64, fileName string) {
	hashMedia, ok := r.pHashTable[hashType]
	if !ok {
		hashMedia = make(map[uint64]string)
		r.pHashTable[hashType] = hashMedia
	}
	hashMedia[hashValue] = fileName
}

func (r *mediaRepositoryMem) FindByPerceptualHash(hashType hash.Type, hashValue uint64, distance int) (int, string) {
	for lHash, fName := range r.pHashTable[hashType] {
		dist, err := hash.Distance(lHash, hashValue)
		if err != nil {
			return -1, ""
		}

		if dist <= distance {
			return dist, fName
		}
	}

	return -1, ""
}

func NewMediaRepositoryMem(dir string, hashTypes []hash.Type) (media.Repository, error) {
	f, err := os.Open(dir)
	if err != nil {
		return nil, err
	}

	entries, _ := f.Readdir(-1)
	if len(entries) <= 0 {
		return nil, errors.New("images/videos not found")
	}

	repository := &mediaRepositoryMem{
		hashTable:  make(map[hash.Type]map[string]string),
		pHashTable: make(map[hash.Type]map[uint64]string),
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		path := filepath.Join(dir, entry.Name())
		m, err := media.NewMedia(path, hashTypes)
		if err != nil {
			return nil, err
		}

		for _, typeHash := range hashTypes {
			switch typeHash {
			case hash.SHA1:
				if h := m.SHA1(); h != "" {
					repository.AppendHash(hash.SHA1, h, m.Name())
				}
			case hash.ED2K:
				if h := m.ED2K(); h != "" {
					repository.AppendHash(hash.ED2K, h, m.Name())
				}
			case hash.AHash:
				if h := m.AHash(); h > 0 {
					repository.AppendPerceptualHash(hash.AHash, h, m.Name())
				}
			case hash.DHash:
				if h := m.DHash(); h > 0 {
					repository.AppendPerceptualHash(hash.DHash, h, m.Name())
				}
			case hash.DHashV:
				if h := m.DHashV(); h > 0 {
					repository.AppendPerceptualHash(hash.DHashV, h, m.Name())
				}
			case hash.DHashD:
				if h := m.DHashD(); h > 0 {
					repository.AppendPerceptualHash(hash.DHashD, h, m.Name())
				}
			case hash.PHash:
				if h := m.PHash(); h > 0 {
					repository.AppendPerceptualHash(hash.PHash, h, m.Name())
				}
			case hash.WHash:
				return nil, errors.New("w-hash not implemented")
			default:
				return nil, errors.New("invalid hash")
			}
		}
	}

	return repository, nil
}
