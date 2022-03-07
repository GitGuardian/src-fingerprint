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
	// IncludeForkedRepos will include fork repositories in fingerprints computation
	// This is available for GitLab and GitHub providers only.
	IncludeForkedRepos bool
	// IncludeArchivedRepos will include archived repositories in fingerprints computation
	// This is only available for GitHub provider only.
	IncludeArchivedRepos bool
	// IncludePublicRepos will include public repositories in fingerprints computation
	// This is only available for GitHub provider only.
	IncludePublicRepos bool
	// Repository private status to display in the output if the provider is 'repository'
	RespositoryIsPrivate bool
	// Use SSH to clone repositories.
	SSHCloning bool
	// BaseURL is the base URL of the API
	BaseURL string
	// Repository name to display in the output if the provider is 'repository'
	RepositoryName string
}
