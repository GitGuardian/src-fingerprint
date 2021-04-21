package gitlab

import (
	"dnacollector"
	"errors"
	"fmt"
	"strings"
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

func (p *Provider) gatherAccessiblePage(page int) ([]dnacollector.GitRepository, int, error) {
	opt := &gitlab.ListProjectsOptions{
		ListOptions: gitlab.ListOptions{
			PerPage: reposPerPage, Page: page,
		}, Statistics: gitlab.Bool(true),
	}

	log.Infof("Gathering page %v for %v\n", page, p.client.BaseURL())

	repos, resp, err := p.client.Projects.ListProjects(opt)
	if err != nil {
		return nil, 0, err
	}

	repositories := make([]dnacollector.GitRepository, 0, len(repos))

	for _, repo := range repos {
		if p.options.OmitForks && repo.ForkedFromProject != nil {
			continue
		}

		repositories = append(repositories, createFromGitlabRepo(repo))
	}

	return repositories, resp.TotalPages, nil
}

func (p *Provider) gatherGroupProjectPage(groupID, page int) ([]dnacollector.GitRepository, int, error) {
	opt := &gitlab.ListGroupProjectsOptions{
		ListOptions: gitlab.ListOptions{
			PerPage: reposPerPage,
			Page:    page,
		},
	}

	log.Infof("Gathering page %v for %v\n", page, p.client.BaseURL())

	repos, resp, err := p.client.Groups.ListGroupProjects(groupID, opt)
	if err != nil {
		return nil, 0, err
	}

	repositories := make([]dnacollector.GitRepository, 0, len(repos))

	for _, repo := range repos {
		if p.options.OmitForks && repo.ForkedFromProject != nil {
			continue
		}

		repositories = append(repositories, createFromGitlabRepo(repo))
	}

	return repositories, resp.TotalPages, nil
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

	for _, group := range groups {
		if strings.EqualFold(group.FullPath, name) {
			return group.ID, nil
		}
	}

	return 0, ErrGroupNotFound
}

// Gather gathers user's repositories for the configured token.
func (p *Provider) Gather(user string) ([]dnacollector.GitRepository, error) {
	log.Debugf("Gathering repositories for user %s\n", user)

	// repositories protected by mu, since multiple goroutines will access it
	repositories := make([]dnacollector.GitRepository, 0)
	if user != "" {
		repositories = p.collectFromGroup(repositories, user)
	} else {
		repositories = p.collectAllAccessible(repositories)
	}

	return repositories, nil
}

func (p *Provider) collectAllAccessible(
	repositories []dnacollector.GitRepository) []dnacollector.GitRepository {
	wg := sync.WaitGroup{}

	var mu sync.Mutex

	_, totalPages, err := p.gatherAccessiblePage(1)
	if err != nil {
		log.Errorf("Error gathering first page: %v\n", err)

		return repositories
	}

	for pageCount := 1; pageCount <= totalPages; pageCount++ {
		wg.Add(1)

		go func(page int) {
			defer wg.Done()

			pageRepositories, _, err := p.gatherAccessiblePage(page)
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

	return repositories
}

func (p *Provider) collectFromGroup(repositories []dnacollector.GitRepository,
	user string) []dnacollector.GitRepository {
	groupID, err := p.findGroup(user)
	if err != nil {
		log.Errorf("Error finding group: %v\n", err)

		return repositories
	}

	wg := sync.WaitGroup{}

	var mu sync.Mutex

	_, totalPages, err := p.gatherGroupProjectPage(groupID, 1)
	if err != nil {
		log.Errorf("Error gathering first page: %v\n", err)

		return repositories
	}

	for pageCount := 1; pageCount <= totalPages; pageCount++ {
		wg.Add(1)

		go func(page int) {
			defer wg.Done()

			pageRepositories, _, err := p.gatherGroupProjectPage(groupID, page)
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

	return repositories
}

// CloneRepository clones a Gitlab repository given the token. The token must have the `read_repository` rights.
func (p *Provider) CloneRepository(
	cloner dnacollector.Cloner,
	repository dnacollector.GitRepository) (*git.Repository, error) {
	auth := &http.BasicAuth{
		Username: p.token,
		Password: p.token,
	}

	// If token doesn't exist, don't try to basic auth
	if p.token == "" {
		auth = nil
	}

	return cloner.CloneRepository(repository.GetHTTPUrl(), auth)
}
