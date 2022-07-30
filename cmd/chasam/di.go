package main

import (
	"github.com/tsmweb/chasam/app/hash"
	"github.com/tsmweb/chasam/app/media"
	"github.com/tsmweb/chasam/infra/repository"
)

type Provider struct {
	mediaRepositoryMem media.Repository
}

func CreateProvider() *Provider {
	return &Provider{}
}

func (p *Provider) MediaRepositoryMem(dir string, hashTypes []hash.Type) (media.Repository, error) {
	if p.mediaRepositoryMem == nil {
		repo, err := repository.NewMediaRepositoryMem(dir, hashTypes)
		if err != nil {
			return nil, err
		}
		p.mediaRepositoryMem = repo
	}
	return p.mediaRepositoryMem, nil
}
