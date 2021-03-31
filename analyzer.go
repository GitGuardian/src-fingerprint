package dnacollector

import (
	git "gopkg.in/src-d/go-git.v4/plumbing/object"
)

// Analyzer analyzer a commit to extract its author.
type Analyzer struct {
}

// AnalyzeCommit extracts author and commiter from a commit.
func (a *Analyzer) AnalyzeCommit(commit *git.Commit) (author, commiter git.Signature) {
	return commit.Author, commit.Committer
}
