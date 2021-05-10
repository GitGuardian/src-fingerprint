package cloner

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	log "github.com/sirupsen/logrus"
)

const gitExitUnclean = 128

// Cloner represents a cloner of git repository.
type Cloner interface {
	CloneRepository(url string) (string, error)
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
	}

	if cacheDir, err := os.UserCacheDir(); err == nil {
		cacheDir = filepath.Join(cacheDir, "srcfingerprint")
		diskCloner.BaseDir = cacheDir

		return diskCloner
	}

	return diskCloner
}

// CloneRepository clones a git repository given its information.
func (d *DiskCloner) CloneRepository(url string) (string, error) {
	tmpDir, err := os.MkdirTemp(d.BaseDir, "srcfingerprint-")
	if err != nil {
		return "", err
	}

	if err := cloneGitRepository(tmpDir, url); err != nil {
		os.RemoveAll(tmpDir)

		return "", err
	}

	return tmpDir, nil
}

func cloneGitRepository(destDir, gitRepoURL string) error {
	var outbuf, errbuf bytes.Buffer
	// git clone github.com/author/name.git /tmp/workdir/author-name/clone
	cmd := exec.Command("git", "clone", gitRepoURL, destDir)
	cmd.Stdout = &outbuf
	cmd.Stderr = &errbuf

	cmd.Env = append(os.Environ(), "GIT_TERMINAL_PROMPT=0")

	err := cmd.Run()
	if exitError, ok := err.(*exec.ExitError); ok {
		stderr := strings.TrimSpace(errbuf.String())

		if exitError.ExitCode() == gitExitUnclean {
			log.WithError(err).WithFields(log.Fields{
				"op":     "gitError",
				"stderr": stderr,
			}).WithField("url", gitRepoURL).Warnf("missing repo")
		} else {
			log.WithError(err).WithFields(log.Fields{
				"op":     "gitError",
				"stderr": stderr,
			}).WithField("url", gitRepoURL).Errorf("unhandled git error")
		}

		return errors.New("")
	}

	return err
}
