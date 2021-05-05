package provider

import (
	"errors"
	"srcfingerprint/cloner"
	"time"

	git "gopkg.in/src-d/go-git.v4"
)

// Generic Repository Structure.
type Repository struct {
	name        string
	sshURL      string
	httpURL     string
	createdAt   time.Time
	storageSize int64
}

// GetName returns the name of the repository.
func (r *Repository) GetName() string { return r.name }

// GetSSHUrl returns the SSH URL of the repository.
func (r *Repository) GetSSHUrl() string { return r.sshURL }

// GetHTTPUrl returns the HTTP URL of the repository.
func (r *Repository) GetHTTPUrl() string { return r.httpURL }

// GetCreatedAt returns the creation time of the repository.
func (r *Repository) GetCreatedAt() time.Time { return r.createdAt }

// GetStorageSize returns the storage size of the repository.
func (r *Repository) GetStorageSize() int64 { return r.storageSize }

type GenericProvider struct {
}

func NewGenericProvider(options Options) Provider {
	return &GenericProvider{}
}

func (p *GenericProvider) Gather(user string) ([]GitRepository, error) {
	if user == "" {
		return nil, errors.New("This provider requires a object. Example: src-fingerprint -p repository -o 'git@github.com:GitGuardian/gg-shield.git'") // nolint
	}

	return []GitRepository{&Repository{
		name:        "",
		httpURL:     user,
		createdAt:   time.Time{},
		storageSize: 0,
	}}, nil
}

func (p *GenericProvider) CloneRepository(
	cloner cloner.Cloner,
	repository GitRepository) (*git.Repository, error) {
	return cloner.CloneRepository(repository.GetHTTPUrl(), nil)
}
