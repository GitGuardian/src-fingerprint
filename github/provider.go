package github

import (
	"context"
	"fmt"
	"net/url"
	"sync"
	"time"

	"github.com/google/go-github/github"
	log "github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
	git "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing/transport/http"

	"dnacollector"
)

const (
	// DefaultGithubAPIURL is the default API URL
	DefaultGithubAPIURL = "https://api.github.com/"
)

// Repository represents a Github repository
type Repository struct {
	name        string
	sshURL      string
	httpURL     string
	createdAt   time.Time
	storageSize int64
}

// GetName returns the name of the repository
func (r Repository) GetName() string { return r.name }

// GetSSHUrl returns the SSH URL of the repository
func (r Repository) GetSSHUrl() string { return r.sshURL }

// GetHTTPUrl returns the HTTP URL of the repository
func (r Repository) GetHTTPUrl() string { return r.httpURL }

// GetCreatedAt returns the creation time of the repository
func (r Repository) GetCreatedAt() time.Time { return r.createdAt }

// GetStorageSize returns the storage size of the repository
func (r Repository) GetStorageSize() int64 { return r.storageSize }

// Provider is capable of gathering Github repositories from an org
type Provider struct {
	client  *github.Client
	options dnacollector.ProviderOptions
	token   string
}

func createFromGithubRepo(r *github.Repository) *Repository {
	return &Repository{
		name:        r.GetName(),
		sshURL:      r.GetSSHURL(),
		httpURL:     r.GetHTMLURL(),
		createdAt:   r.GetCreatedAt().Time,
		storageSize: int64(r.GetSize()),
	}
}

const reposPerPage = 100

// NewProvider creates a new Github Provider
func NewProvider(token string, options dnacollector.ProviderOptions) *Provider {
	client := github.NewClient(oauth2.NewClient(
		context.TODO(),
		oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token}),
	))
	if options.BaseURL != "" {
		baseParsedURL, err := url.Parse(options.BaseURL)
		if err != nil {
			panic(fmt.Sprintf("Github Base URL is not a valid url: %v", options.BaseURL))
		}
		client.BaseURL = baseParsedURL
	}
	return &Provider{
		client:  client,
		options: options,
		token:   token,
	}
}

func (p *Provider) gatherPage(user string, page int) ([]dnacollector.GitRepository, error) {
	opt := &github.RepositoryListByOrgOptions{
		ListOptions: github.ListOptions{
			PerPage: reposPerPage, Page: page,
		},
	}

	log.Infof("Gathering page %v for %s\n", page, user)

	repos, _, err := p.client.Repositories.ListByOrg(context.Background(), user, opt)
	if err != nil {
		return nil, err
	}

	repositories := make([]dnacollector.GitRepository, 0, len(repos))
	for _, repo := range repos {
		if p.options.OmitForks && repo.GetFork() {
			continue
		}
		repositories = append(repositories, createFromGithubRepo(repo))
	}
	return repositories, nil
}

// Gather gather user's git repositories and send them to outputChannel
func (p *Provider) Gather(user string) ([]dnacollector.GitRepository, error) {
	log.Debugf("Gathering repositories for Github org %s\n", user)

	opt := &github.RepositoryListByOrgOptions{
		ListOptions: github.ListOptions{
			PerPage: reposPerPage, Page: 1,
		},
	}
	repos, resp, err := p.client.Repositories.ListByOrg(context.Background(), user, opt)
	if err != nil {
		return nil, err
	}

	pagesCount := resp.LastPage

	log.Infof("Gathering %v pages for %s\n", pagesCount, user)
	wg := sync.WaitGroup{}

	var (
		mu sync.Mutex
		// repositories protected by mu, since multiple goroutines will access it
		repositories []dnacollector.GitRepository
	)

	repositories = make([]dnacollector.GitRepository, 0, pagesCount*reposPerPage)
	for _, repo := range repos {
		repositories = append(repositories, createFromGithubRepo(repo))
	}
	for pageCount := 1; pageCount <= pagesCount; pageCount++ {
		wg.Add(1)
		go func(page int) {
			defer wg.Done()

			pageRepositories, err := p.gatherPage(user, page)
			if err != nil {
				log.Errorf("Error gathering page %v:%v\n", page, err)
				return
			}

			mu.Lock()
			repositories = append(repositories, pageRepositories...)
			mu.Unlock()
		}(pageCount)
	}

	wg.Wait()

	return repositories, nil
}

// CloneRepository clones a Github repository given the token. The token must have the `read_repository` rights
func (p *Provider) CloneRepository(cloner dnacollector.Cloner, repository dnacollector.GitRepository) (*git.Repository, error) {
	auth := &http.BasicAuth{
		Username: p.token,
		Password: p.token,
	}

	return cloner.CloneRepository(repository.GetHTTPUrl(), auth)
}
