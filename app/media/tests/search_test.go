package tests

import (
	"context"
	"fmt"
	"github.com/tsmweb/chasam/app/hash"
	"github.com/tsmweb/chasam/app/media"
	"github.com/tsmweb/chasam/infra/repository"
	"os"
	"runtime"
	"testing"
)

var (
	repo media.Repository
)

func TestSearch(t *testing.T) {
	source := "/home/martins/Downloads/images/search"
	target := "/home/martins/Downloads/images/benchmark"
	hashTypes := []hash.Type{
		hash.SHA1,
		hash.ED2K,
		hash.AHash,
		hash.DHash,
		hash.DHashV,
		hash.PHash,
	}

	_repo, err := repository.NewMediaRepositoryMem(source, hashTypes)
	if err != nil {
		t.Error(err)
	}
	repo = _repo

	s := media.NewSearch(
		context.Background(),
		target,
		hashTypes,
		onError,
		onSearch,
		onMatch,
		runtime.NumCPU())
	s.Run()
}

func onSearch(_ context.Context, m *media.Media) (bool, error) {
	if m.Type() != "image" {
		return false, nil
	}

	if src := repo.FindByHash(hash.SHA1, m.SHA1()); src != "-1" {
		m.AddMatch(src, hash.SHA1.String(), 0)
		return true, nil
	}

	if src := repo.FindByHash(hash.ED2K, m.ED2K()); src != "-1" {
		m.AddMatch(src, hash.ED2K.String(), 0)
		return true, nil
	}

	hamming := 1

	if dist, src := repo.FindByPerceptualHash(hash.AHash, m.AHash(), hamming); dist != -1 {
		m.AddMatch(src, hash.AHash.String(), dist)
		return true, nil
	}

	if dist, src := repo.FindByPerceptualHash(hash.DHash, m.DHash(), hamming); dist != -1 {
		m.AddMatch(src, hash.DHash.String(), dist)
		return true, nil
	}

	if dist, src := repo.FindByPerceptualHash(hash.DHashV, m.DHashV(), hamming); dist != -1 {
		m.AddMatch(src, hash.DHashV.String(), dist)
		return true, nil
	}

	if dist, src := repo.FindByPerceptualHash(hash.PHash, m.PHash(), hamming); dist != -1 {
		m.AddMatch(src, hash.PHash.String(), dist)
		return true, nil
	}
	return false, nil
}

func onMatch(_ context.Context, m *media.Media) {
	fmt.Printf("Name: %s\n", m.Name())
	fmt.Printf("Path: %s\n", m.Path())
	fmt.Printf("Type: %s\n", m.Type())
	fmt.Printf("SHA1: %s\n", m.SHA1())
	fmt.Printf("ED2K: %s\n", m.ED2K())
	fmt.Printf("Ahash: %s\n", hash.FormatToHex(m.AHash()))
	fmt.Printf("Dhash: %s\n", hash.FormatToHex(m.DHash()))
	fmt.Printf("DhashV: %s\n", hash.FormatToHex(m.DHashV()))
	fmt.Printf("Phash: %s\n", hash.FormatToHex(m.PHash()))
	fmt.Printf("Match: %v\n\n", m.Match())
}

func onError(_ context.Context, err error) {
	fmt.Fprintln(os.Stderr, err.Error())
}
