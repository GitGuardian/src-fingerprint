package provider

import (
	"time"

	"srcfingerprint/cloner"
)

// GitRepository represents a git repository for the Extractor.
type GitRepository interface {
	// GetName is the name of the repository.
	GetName() string

	// GetSSHUrl is the SSH Url of the repository.
	GetSSHUrl() string

	// GetHTTPUrl is the HTTP Url of the repository.
	GetHTTPUrl() string

	// GetCreatedAt is the time of creation of the repository
	GetCreatedAt() time.Time

	// GetStorageSiwe is the size of the repository
	GetStorageSize() int64

	// GetPrivate returns either the repository is private or not.
	GetPrivate() bool
}

// Provider is the interface to implement for a Git provider.
type Provider interface {
	// Gather is the function gathering git repositories given an user
	// from the provider.
	Gather(user string) ([]GitRepository, error)

	CloneRepository(cloner cloner.Cloner, repository GitRepository) (string, error)
}

// Options represents options for the Provider.
type Options struct {
	// OmitForks will tell the Provider to omit fork repositories
	OmitForks bool
	// SkipArchived will skip archived repositories
	SkipArchived bool
	// BaseURL is the base URL of the API
	BaseURL string
}
