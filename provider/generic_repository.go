package provider

import (
	"errors"
	"os"
	"path/filepath"
	"srcfingerprint/cloner"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
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
		return nil, errors.New(
			"this provider requires an object. " +
				"Example: src-fingerprint -p repository -u 'https://github.com/GitGuardian/src-fingerprint'",
		)
	}

	var name string
	if p.options.RepositoryName != "" {
		name = p.options.RepositoryName
	} else {
		path := user
		if _, err := os.Stat(user); err == nil || !os.IsNotExist(err) {
			if absPath, err := filepath.Abs(user); err == nil {
				log.Debugf("`%v` is a local path, will use absolute path `%v`", user, absPath)
				path = absPath
			}
		}

		// Split the repository URL or patch and use the last part
		parts := strings.Split(path, "/")
		if parts[len(parts)-1] == ".git" && len(parts) > 2 {
			// As "path/to/project/.git" is valid, we use the second to last part when the last part is ".git"
			name = parts[len(parts)-2]
		} else {
			name = parts[len(parts)-1]
			name = strings.TrimSuffix(name, ".git")
		}

		log.Warnf("Name of the repository unspecified (see --repo-name). '%v' has been inferred from the object.", name)
	}

	return []GitRepository{&Repository{
		name:        name,
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
