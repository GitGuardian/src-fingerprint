package srcfingerprint

import (
	"srcfingerprint/cloner"
	"srcfingerprint/extractor"
	"srcfingerprint/provider"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
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

type ExtractorMock struct {
	gitFiles []extractor.GitFile
	i        *int
}

func (m ExtractorMock) Next() (*extractor.GitFile, bool) {
	if *m.i < len(m.gitFiles) {
		*m.i++
		return &m.gitFiles[*m.i-1], true
	} else {
		return nil, false
	}
}

func (m ExtractorMock) Run(string, string) {}

type ExtractorMockMaker struct {
	gitFiles []extractor.GitFile
}

func (e ExtractorMockMaker) Make() extractor.Extractor {
	i := new(int)
	*i = 0
	return ExtractorMock{e.gitFiles, i}
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
	eventChan := make(chan PipelineEvent)
	providerMock := &ProviderMock{}
	repository := createGitRepository("repoName")

	gitFile := extractor.GitFile{
		Sha:      "a_sha",
		Type:     "a_type",
		Filepath: "a_filepath",
		Size:     "a_size",
	}
	extractorMockMaker := &ExtractorMockMaker{[]extractor.GitFile{gitFile}}
	pipeline := Pipeline{Provider: providerMock, ExtractorMaker: extractorMockMaker}

	gitRepository := "git://host/repoName"

	providerMock.On("CloneRepository", nil, repository).Return(gitRepository, nil)

	go func() {
		defer close(eventChan)

		err := pipeline.ExtractRepository(repository, "", eventChan)
		assert.Nil(suite.T(), err, "ExtractRepository returned an error")
	}()

	events := make([]PipelineEvent, 0)
	for event := range eventChan {
		events = append(events, event)
	}

	expectedEvents := []PipelineEvent{
		ResultGitFilePipelineEvent{
			Repository: repository,
			GitFile:    &gitFile,
		},
		RepositoryPipelineEvent{true, true, "repoName"},
	}

	providerMock.AssertExpectations(suite.T())
	assert.Equal(suite.T(), expectedEvents, events)
}

func (suite *PipelineTestSuite) TestExtractRepositories() {
	eventChan := make(chan PipelineEvent)
	providerMock := &ProviderMock{}
	repository := createGitRepository("repoName")

	gitFile := extractor.GitFile{
		Sha:      "a_sha",
		Type:     "a_type",
		Filepath: "a_filepath",
		Size:     "a_size",
	}
	extractorMockMaker := &ExtractorMockMaker{[]extractor.GitFile{gitFile}}
	pipeline := Pipeline{Provider: providerMock, ExtractorMaker: extractorMockMaker}

	gitRepository := "git://host/repoName"

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
		ResultGitFilePipelineEvent{
			Repository: repository,
			GitFile:    &gitFile,
		},
		RepositoryPipelineEvent{true, true, "repoName"},
	}

	providerMock.AssertExpectations(suite.T())
	assert.Equal(suite.T(), expectedEvents, events)
}

func TestPipeline(t *testing.T) {
	suite.Run(t, new(PipelineTestSuite))
}
