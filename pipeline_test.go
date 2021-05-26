package srcfingerprint

import (
	"path/filepath"
	"srcfingerprint/cloner"
	"srcfingerprint/provider"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	git "gopkg.in/src-d/go-git.v4"
)

type PipelineTestSuite struct {
	suite.Suite
}

type ProviderMock struct {
	mock.Mock
}

func (mock *ProviderMock) Gather(user string) ([]provider.GitRepository, error) {
	args := mock.Called(user)

	return args.Get(0).([]provider.GitRepository), args.Error(1)
}

func (mock *ProviderMock) CloneRepository(cloner cloner.Cloner, repository provider.GitRepository) (string, error) {
	args := mock.Called(cloner, repository)

	return args.String(0), args.Error(1)
}

type gitRepositoryMock struct{ name string }

func (m gitRepositoryMock) GetName() string         { return m.name }
func (m gitRepositoryMock) GetSSHUrl() string       { return "" }
func (m gitRepositoryMock) GetHTTPUrl() string      { return "" }
func (m gitRepositoryMock) GetCreatedAt() time.Time { return time.Unix(0, 0) }
func (m gitRepositoryMock) GetStorageSize() int64   { return 0 }
func (m gitRepositoryMock) GetPrivate() bool        { return true }

func createGitRepository(name string) provider.GitRepository {
	return gitRepositoryMock{name: name}
}

func openTestGitRepository(t *testing.T) *git.Repository {
	repopath, _ := filepath.Abs(filepath.Join("testdata", "gitrepo", "git"))
	repository, err := git.PlainOpen(repopath)
	if err != nil {
		t.Fatalf("could not open test git repository: %v", err)
	}
	return repository
}

func (suite *PipelineTestSuite) TestGather() {
	outputChan := make(chan provider.GitRepository)
	wg := &sync.WaitGroup{}
	providerMock := &ProviderMock{}
	pipeline := Pipeline{
		Provider: providerMock,
	}

	providerMock.On("Gather", "user").Return([]provider.GitRepository{createGitRepository("1")}, nil)

	wg.Add(1)
	go pipeline.gather(wg, nil, "user", outputChan)

	repositories := make([]provider.GitRepository, 0, 2)
	for output := range outputChan {
		repositories = append(repositories, output)
	}
	wg.Wait()

	providerMock.AssertExpectations(suite.T())
	assert.Equal(suite.T(), []provider.GitRepository{gitRepositoryMock{name: "1"}}, repositories)
}

func (suite *PipelineTestSuite) TestExtractGitRepository() {
	suite.T().Skip("Skip until repository is stable")
	eventChan := make(chan PipelineEvent)
	provider := &ProviderMock{}
	repository := createGitRepository("repoName")
	pipeline := Pipeline{Provider: provider}

	gitRepository := openTestGitRepository(suite.T())
	commitIter, _ := gitRepository.CommitObjects()
	// firstCommit, _ := commitIter.Next()
	commitIter.Close()

	provider.On("CloneRepository", nil, repository).Return(gitRepository, nil)

	go func() {
		defer close(eventChan)

		pipeline.ExtractRepository(repository, "", eventChan)
	}()

	events := make([]PipelineEvent, 0)
	for event := range eventChan {
		events = append(events, event)
	}

	expectedEvents := []PipelineEvent{
		// ResultPipelineEvent{
		// 	Repository: repository,
		// 	Commit:     firstCommit,
		// 	Author:     firstCommit.Author,
		// 	Committer:  firstCommit.Committer,
		// },
		RepositoryPipelineEvent{true, true, "repoName"},
	}

	provider.AssertExpectations(suite.T())
	assert.Equal(suite.T(), expectedEvents, events)
}

func (suite *PipelineTestSuite) TestExtractRepositories() {
	suite.T().Skip("Skip until repository is stable") // Skip for now
	eventChan := make(chan PipelineEvent)
	providerMock := &ProviderMock{}
	repository := createGitRepository("repoName")
	pipeline := Pipeline{Provider: providerMock}

	gitRepository := openTestGitRepository(suite.T())
	commitIter, _ := gitRepository.CommitObjects()
	// firstCommit, _ := commitIter.Next()
	commitIter.Close()

	providerMock.On("Gather", "user").Return([]provider.GitRepository{repository}, nil)
	providerMock.On("CloneRepository", nil, repository).Return(gitRepository, nil)

	go func() {
		defer close(eventChan)

		pipeline.ExtractRepositories("user", "", eventChan)
	}()

	events := make([]PipelineEvent, 0)
	for event := range eventChan {
		events = append(events, event)
	}

	expectedEvents := []PipelineEvent{
		RepositoryListPipelineEvent{Repositories: []provider.GitRepository{repository}},
		// ResultPipelineEvent{
		// 	Repository: repository,
		// 	Commit:     firstCommit,
		// 	Author:     firstCommit.Author,
		// 	Committer:  firstCommit.Committer,
		// },
		RepositoryPipelineEvent{true, true, "repoName"},
	}

	providerMock.AssertExpectations(suite.T())
	assert.Equal(suite.T(), expectedEvents, events)
}

func TestPipeline(t *testing.T) {
	suite.Run(t, new(PipelineTestSuite))
}
