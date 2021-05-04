package srcfingerprint

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
)

type AnalyzerTestSuite struct {
	suite.Suite
}

func (suite *AnalyzerTestSuite) TestAnalyzeCommit() {
	analyzer := Analyzer{}
	commit := &object.Commit{
		Author: object.Signature{
			Name:  "Author",
			Email: "author@example.com",
		},
		Committer: object.Signature{
			Name:  "Committer",
			Email: "committer@example.com",
		},
	}

	author, committer := analyzer.AnalyzeCommit(commit)

	assert.Equal(suite.T(), object.Signature{Name: "Author", Email: "author@example.com"}, author)
	assert.Equal(suite.T(), object.Signature{Name: "Committer", Email: "committer@example.com"}, committer)
}

func TestAnalyzer(t *testing.T) {
	suite.Run(t, new(AnalyzerTestSuite))
}
