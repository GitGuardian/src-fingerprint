package cloner

import (
	"context"
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"
)

const gitExitUnclean = 128

// Cloner represents a cloner of git repository.
type Cloner interface {
	CloneRepository(ctx context.Context, url string) (string, error)
}

// DiskCloner closes a git repository on disk in a temporary file.
type DiskCloner struct {
	BaseDir string
}

// NewDiskCloner creates a new DiskCloner.
// If baseDir is an empty string the default user cache folder + "/srcfingerprint"
// will be used, if this is not available /tmp will be used.
func NewDiskCloner(baseDir string) *DiskCloner {
	diskCloner := &DiskCloner{BaseDir: "/tmp"}

	if baseDir != "" {
		if _, err := os.Stat(baseDir); !os.IsNotExist(err) {
			diskCloner.BaseDir = baseDir

			return diskCloner
		}

		log.Warnln("Clone directory provided does not exist locally. Will clone in default cache directory.")
	}

	if cacheDir, err := os.UserCacheDir(); err == nil {
		cacheDir = filepath.Join(cacheDir, "srcfingerprint")
		if err := os.MkdirAll(cacheDir, os.ModePerm); err == nil {
			diskCloner.BaseDir = cacheDir
		}

		return diskCloner
	}

	return diskCloner
}

// CloneRepository clones a git repository given its information.
func (d *DiskCloner) CloneRepository(ctx context.Context, url string) (string, error) {
	tmpDir, err := os.MkdirTemp(d.BaseDir, "srcfingerprint-")
	if err != nil {
		return "", err
	}

	if err := cloneGitRepository(ctx, tmpDir, url); err != nil {
		os.RemoveAll(tmpDir)

		return "", err
	}

	return tmpDir, nil
}
