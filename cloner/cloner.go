package cloner

import (
	"io/ioutil"

	"gopkg.in/src-d/go-billy.v4/osfs"
	git "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing/cache"
	"gopkg.in/src-d/go-git.v4/plumbing/transport"
	"gopkg.in/src-d/go-git.v4/storage/filesystem"
)

// Cloner represents a cloner of git repository.
type Cloner interface {
	CloneRepository(url string, auth transport.AuthMethod) (*git.Repository, error)
}

// DiskCloner closes a git repository on disk in a temporary file.
type DiskCloner struct{}

// CloneRepository clones a git repository given its information.
func (*DiskCloner) CloneRepository(url string, auth transport.AuthMethod) (*git.Repository, error) {
	tmpDir, err := ioutil.TempDir("", "fs-")
	if err != nil {
		return nil, err
	}

	fs := osfs.New(tmpDir)

	return git.Clone(filesystem.NewStorage(fs, cache.NewObjectLRUDefault()), nil, &git.CloneOptions{
		URL:      url,
		Progress: ioutil.Discard,
		Auth:     auth,
	})
}
