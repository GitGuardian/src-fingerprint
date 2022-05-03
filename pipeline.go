package srcfingerprint

import (
	"context"
	"errors"
	"srcfingerprint/cloner"
	"srcfingerprint/provider"
	"sync"
	"time"

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
	object string,
	output chan<- provider.GitRepository,
	limit int) {
	defer wg.Done()
	defer close(output)

	repositories, err := p.Provider.Gather(object)
	if err != nil {
		log.Errorf("Gathering repositories failed: %v\n", err)

		return
	}

	p.publishEvent(eventChan, RepositoryListPipelineEvent{repositories})

	collected := 0
	ignored := 0

	for index, repository := range repositories {
		if limit > 0 && index >= limit {
			ignored++
		} else {
			collected++
			output <- repository
		}
	}

	if ignored > 0 {
		log.Warnln("Limit reached for number of repositories")
		log.Warnf("Collected %d repos, ignored %d repos.", collected, ignored)
	}

	log.Infoln("Done gathering repositories")
}

// ExtractRepository extracts for a single repository.
func (p *Pipeline) ExtractRepository(ctx context.Context, repository provider.GitRepository, after string, eventChan chan<- PipelineEvent) error { // nolint
	defer p.publishEvent(eventChan, RepositoryPipelineEvent{true, repository.GetPrivate(), repository.GetName()})

	log.Infof("Cloning repo %v\n", repository.GetName())

	gitRepository, err := p.Provider.CloneRepository(ctx, p.Cloner, repository)
	if err != nil {
		return err
	}

	log.Infof("Cloned repo %v (size: %v KB)\n", repository.GetName(), repository.GetStorageSize())

	extractorGitFile := NewFastExtractor()
	extractorGitFile.Run(gitRepository, after)

loop:
	for {
		select {
		case gitFile, opened := <-extractorGitFile.ChanGitFiles:
			if !opened {
				break loop
			}
			p.publishEvent(eventChan, ResultGitFilePipelineEvent{repository, gitFile})
		case <-ctx.Done():
			return errors.New("timeout reached while extracting files")
		}
	}

	log.Infof("Done extracting %v\n", repository.GetName())

	return nil
}

const (
	defaultExtractionWorkersCount = 10
)

// ExtractRepositories extract repositories and analyze it for a given user and provider.
func (p *Pipeline) ExtractRepositories(
	object string,
	after string,
	eventChan chan<- PipelineEvent,
	limit int,
	timeout time.Duration,
) {
	if object != "" {
		log.Infof("Extracting all repositories for the org or group %v and accessible with the specified token.\n", object)
	} else {
		log.Info("Extracting all repositories accessible to the user providing the token.")
	}

	repositoryChannel := make(chan provider.GitRepository)

	extractionWorkersCount := p.ClonersCount
	if extractionWorkersCount == 0 {
		extractionWorkersCount = defaultExtractionWorkersCount
	}

	wg := sync.WaitGroup{}

	ctx, cancel := context.WithCancel(context.Background())
	if timeout > 0 {
		ctx, cancel = context.WithTimeout(context.Background(), timeout)
	}

	defer cancel()
	wg.Add(1)

	go p.gather(&wg, eventChan, object, repositoryChannel, limit)

	for i := 0; i < extractionWorkersCount; i++ {
		wg.Add(1)

		go func(ctx context.Context) {
			defer wg.Done()

			for repository := range repositoryChannel {
				if err := p.ExtractRepository(ctx, repository, after, eventChan); err != nil {
					log.Errorf("extracting %v failed: %v\n", repository.GetName(), err)
				}
			}
		}(ctx)
	}

	wg.Wait()
	log.Infof("Done extracting the org or group %v\n", object)
}
