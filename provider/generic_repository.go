package provider

import (
	"errors"
	"srcfingerprint/cloner"
	"time"
)

// Repository Structure. Generic.
type Repository struct {
	name        string
	sshURL      string
	httpURL     string
	createdAt   time.Time
	storageSize int64
	private     bool
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

// GetPrivate returns either the repository is private or not.
func (r *Repository) GetPrivate() bool { return r.private }

type GenericProvider struct {
	options Options
}

func NewGenericProvider(options Options) Provider {
	return &GenericProvider{options}
}

func (p *GenericProvider) Gather(user string) ([]GitRepository, error) {
	if user == "" {
		return nil, errors.New("This provider requires a object. Example: src-fingerprint -p repository -u 'git@github.com:GitGuardian/gg-shield.git'") // nolint
	}

	return []GitRepository{&Repository{
		name:        p.options.RepositoryName,
		httpURL:     user,
		createdAt:   time.Time{},
		storageSize: 0,
		private:     p.options.RespositoryIsPrivate,
	}}, nil
}

func (p *GenericProvider) CloneRepository(
	cloner cloner.Cloner,
	repository GitRepository) (string, error) {
	return cloner.CloneRepository(repository.GetHTTPUrl())
}
