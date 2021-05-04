package provider

import "time"

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
