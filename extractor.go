package dnacollector

import (
	git "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
)

// Extractor extracts commits for a given repository
type Extractor struct {
	iter object.CommitIter
}

// NewExtractor creates a new Extractor
func NewExtractor(repository *git.Repository) (*Extractor, error) {
	iter, err := repository.CommitObjects()
	if err != nil {
		return nil, err
	}
	return &Extractor{
		iter: iter,
	}, nil
}

// ExtractNextCommit extract the next  commit of the repository.
// It returns an io.EOF error if there are no more commits
func (e *Extractor) ExtractNextCommit() (*object.Commit, error) {
	return e.iter.Next()
}
