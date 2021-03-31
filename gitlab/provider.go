package gitlab

import (
	"dnacollector"
	"errors"
	"fmt"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
	gitlab "github.com/xanzy/go-gitlab"
	git "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing/transport/http"
)

const (
	// DefaultGithubAPIURL is the default API URL.
	DefaultGitLabAPIURL = "https://gitlab.com/api/v4"
)

// Repository represents a Gitlab repository.
type Repository struct {
	name        string
	sshURL      string
	httpURL     string
	createdAt   time.Time
	storageSize int64
}

// ErrGroupNotFound is the error returned when group can not be found.
var ErrGroupNotFound = errors.New("group not found")

// GetName returns the name of the repository.
func (r *Repository) GetName() string { return r.name }

// GetSSHUrl returns the SSH URL of the repository.
func (r *Repository) GetSSHUrl() string { return r.sshURL }

// GetHTTPUrl returns the HTTP URL of the repository.
func (r *Repository) GetHTTPUrl() string { return r.httpURL }

// GetCreatedAt returns the creation time of the repository.
func (r *Repository) GetCreatedAt() time.Time { return r.createdAt }

// GetStorageSize returns the storage size of the repository.
func (r *Repository) GetStorageSize() int64 { return r.storageSize }

// Provider represents a Gitlab Provider. It can gather the list of repositories a given user.
type Provider struct {
	token   string
	client  *gitlab.Client
	options dnacollector.ProviderOptions
}

// NewProvider  creates a Provider given a token.
// If accessing private repositories, token must not be empty.
func NewProvider(token string, options dnacollector.ProviderOptions) *Provider {
	GitLabBaseURL := DefaultGitLabAPIURL
	if options.BaseURL != "" {
		GitLabBaseURL = options.BaseURL
	}

	client, err := gitlab.NewClient(token, gitlab.WithBaseURL(GitLabBaseURL))
	if err != nil {
		panic(fmt.Sprintf("could not set base URL for gitlab client: %v", err))
	}

	return &Provider{
		token:   token,
		client:  client,
		options: options,
	}
}

const reposPerPage = 100

func createFromGitlabRepo(r *gitlab.Project) *Repository {
	storageSize := int64(0)
	if r.Statistics != nil {
		storageSize = r.Statistics.RepositorySize
	}

	return &Repository{
		name:        r.Name,
		sshURL:      r.SSHURLToRepo,
		httpURL:     r.HTTPURLToRepo,
		createdAt:   *r.CreatedAt,
		storageSize: storageSize,
	}
}

func (p *Provider) gatherPage(page int) ([]dnacollector.GitRepository, error) {
	opt := &gitlab.ListProjectsOptions{
		ListOptions: gitlab.ListOptions{
			PerPage: reposPerPage, Page: page,
		}, Statistics: gitlab.Bool(true),
	}

	log.Infof("Gathering page %v for %v\n", page, p.client.BaseURL())

	repos, _, err := p.client.Projects.ListProjects(opt)
	if err != nil {
		return nil, err
	}

	repositories := make([]dnacollector.GitRepository, 0, len(repos))

	for _, repo := range repos {
		if p.options.OmitForks && repo.ForkedFromProject != nil {
			continue
		}

		repositories = append(repositories, createFromGitlabRepo(repo))
	}

	return repositories, nil
}

func (p *Provider) findGroup(name string) (int, error) {
	groups, _, err := p.client.Groups.ListGroups(&gitlab.ListGroupsOptions{
		Search: &name,
	})
	if err != nil {
		return 0, err
	}

	if len(groups) < 1 {
		return 0, ErrGroupNotFound
	}

	return groups[0].ID, nil
}

// Gather gathers user's repositories for the configured token.
func (p *Provider) Gather(user string) ([]dnacollector.GitRepository, error) {
	log.Debugf("Gathering repositories for user %s\n", user)

	groupID, err := p.findGroup(user)
	if err != nil {
		return nil, err
	}

	opt := &gitlab.ListGroupProjectsOptions{
		ListOptions: gitlab.ListOptions{
			PerPage: reposPerPage,
			Page:    1,
		},
	}

	repos, resp, err := p.client.Groups.ListGroupProjects(groupID, opt)
	if err != nil {
		return nil, err
	}

	pagesCount := resp.TotalPages

	log.Infof("Gathering %v pages for %s\n", pagesCount, user)

	wg := sync.WaitGroup{}

	var mu sync.Mutex

	// repositories protected by mu, since multiple goroutines will access it
	repositories := make([]dnacollector.GitRepository, 0, pagesCount*reposPerPage)
	for _, repo := range repos {
		repositories = append(repositories, createFromGitlabRepo(repo))
	}

	for pageCount := 1; pageCount <= pagesCount; pageCount++ {
		wg.Add(1)

		go func(page int) {
			defer wg.Done()

			pageRepositories, err := p.gatherPage(page)
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

// CloneRepository clones a Gitlab repository given the token. The token must have the `read_repository` rights.
func (p *Provider) CloneRepository(
	cloner dnacollector.Cloner,
	repository dnacollector.GitRepository) (*git.Repository, error) {
	auth := &http.BasicAuth{
		Username: p.token,
		Password: p.token,
	}

	return cloner.CloneRepository(repository.GetHTTPUrl(), auth)
}
