package srcfingerprint

import (
	"srcfingerprint/cloner"
	"srcfingerprint/provider"
	"sync"

	log "github.com/sirupsen/logrus"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
)

// PipelineEvent is the interface for a pipeline event.
type PipelineEvent interface{}

// RepositoryListPipelineEvent is the event fired when the list of repositories has been gathered.
type RepositoryListPipelineEvent struct {
	// Repositories is the list of repositories
	Repositories []provider.GitRepository
}

// ResultCommitPipelineEvent represents the event for a result.
type ResultCommitPipelineEvent struct {
	Repository provider.GitRepository
	Commit     *object.Commit
	Author     object.Signature
	Committer  object.Signature
}

type ResultGitFilePipelineEvent struct {
	Repository provider.GitRepository
	GitFile    *GitFile
}

// RepositoryPipelineEvent represents an event from a repository.
type RepositoryPipelineEvent struct {
	// Finished is true if the given repository is pipeline is done
	Finished bool
	// either the repository is private or not
	Private bool
	// RepositoryName is the name of the repository
	RepositoryName string
}

// CommitPipelineEvent represents an event from a repository.
type CommitPipelineEvent struct {
	// Finished is true if the given repository is pipeline is done
	Finished bool
	// RepositoryName is the name of the repository
	Repository string
}

// Pipeline represents the whole extraction pipeline.
type Pipeline struct {
	Provider provider.Provider
	Cloner   cloner.Cloner
	Analyzer *Analyzer

	ClonersCount int
}

func (p *Pipeline) publishEvent(ch chan<- PipelineEvent, event PipelineEvent) {
	if ch != nil {
		ch <- event
	}
}

// run in its own goroutine.
func (p *Pipeline) gather(
	wg *sync.WaitGroup,
	eventChan chan<- PipelineEvent,
	user string,
	output chan<- provider.GitRepository) {
	defer wg.Done()
	defer close(output)

	repositories, err := p.Provider.Gather(user)
	if err != nil {
		log.Errorf("Gathering repositories failed: %v\n", err)

		return
	}

	p.publishEvent(eventChan, RepositoryListPipelineEvent{repositories})

	for _, repository := range repositories {
		output <- repository
	}

	log.Infoln("Done gathering repositories")
}

// ExtractRepository extracts for a single repository.
func (p *Pipeline) ExtractRepository(repository provider.GitRepository, after string, eventChan chan<- PipelineEvent) error { // nolint
	defer p.publishEvent(eventChan, RepositoryPipelineEvent{true, repository.GetPrivate(), repository.GetName()})

	log.Infof("Cloning repo %v\n", repository.GetName())

	gitRepository, err := p.Provider.CloneRepository(p.Cloner, repository)
	if err != nil {
		return err
	}

	log.Infof("Cloned repo %v (size: %v KB)\n", repository.GetName(), repository.GetStorageSize())

	extractorGitFile := NewFastExtractor()
	extractorGitFile.Run(gitRepository, after)

	for gitFile := range extractorGitFile.ChanGitFiles {
		p.publishEvent(eventChan, ResultGitFilePipelineEvent{repository, gitFile})
	}

	log.Infof("Done extracting %v\n", repository.GetName())

	return nil
}

const (
	defaultExtractionWorkersCount = 10
)

// ExtractRepositories extract repositories and analyze it for a given user and provider.
func (p *Pipeline) ExtractRepositories(user string, after string, eventChan chan<- PipelineEvent) {
	log.Infof("Extracting user %v\n", user)

	repositoryChannel := make(chan provider.GitRepository)

	extractionWorkersCount := p.ClonersCount
	if extractionWorkersCount == 0 {
		extractionWorkersCount = defaultExtractionWorkersCount
	}

	wg := sync.WaitGroup{}

	wg.Add(1)

	go p.gather(&wg, eventChan, user, repositoryChannel)

	for i := 0; i < extractionWorkersCount; i++ {
		wg.Add(1)

		go func() {
			defer wg.Done()

			for repository := range repositoryChannel {
				if err := p.ExtractRepository(repository, after, eventChan); err != nil {
					log.Errorf("extracting %v failed: %v\n", repository.GetName(), err)
				}
			}
		}()
	}

	wg.Wait()
	log.Infof("Done extracting user %v\n", user)
}
