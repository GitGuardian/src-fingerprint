package provider

import (
	"context"
	"fmt"
	"net/url"
	"srcfingerprint/cloner"
	"sync"

	"github.com/google/go-github/github"
	log "github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
	git "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing/transport/http"
)

const (
	// DefaultGithubAPIURL is the default API URL.
	DefaultGithubAPIURL = "https://api.github.com/"
)

// Provider is capable of gathering Github repositories from an org.
type GitHubProvider struct {
	client  *github.Client
	options Options
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

// NewProvider creates a new Github Provider.
func NewGitHubProvider(token string, options Options) Provider {
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

	return &GitHubProvider{
		client:  client,
		options: options,
		token:   token,
	}
}

func (p *GitHubProvider) gatherPage(user string, page int) ([]GitRepository, error) {
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

	repositories := make([]GitRepository, 0, len(repos))

	for _, repo := range repos {
		if p.options.OmitForks && repo.GetFork() {
			continue
		}

		repositories = append(repositories, createFromGithubRepo(repo))
	}

	return repositories, nil
}

// Gather gather user's git repositories and send them to outputChannel.
func (p *GitHubProvider) Gather(user string) ([]GitRepository, error) {
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

	var mu sync.Mutex

	// repositories protected by mu, since multiple goroutines will access it
	repositories := make([]GitRepository, 0, pagesCount*reposPerPage)
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

// CloneRepository clones a Github repository given the token. The token must have the `read_repository` rights.
func (p *GitHubProvider) CloneRepository(cloner cloner.Cloner,
	repository GitRepository) (*git.Repository, error) {
	auth := &http.BasicAuth{
		Username: p.token,
		Password: p.token,
	}

	return cloner.CloneRepository(repository.GetHTTPUrl(), auth)
}
