package provider

import (
	"context"
	"errors"
	"fmt"
	"srcfingerprint/cloner"
	"strings"
	"sync"

	log "github.com/sirupsen/logrus"
	gitlab "github.com/xanzy/go-gitlab"
)

const (
	// DefaultGitLabAPIURL is the default API URL.
	DefaultGitLabAPIURL = "https://gitlab.com/api/v4"
)

// ErrGroupNotFound is the error returned when group can not be found.
var ErrGroupNotFound = errors.New("group not found")

// GitLabProvider represents a Gitlab Provider. It can gather the list of repositories a given user.
type GitLabProvider struct {
	token   string
	client  *gitlab.Client
	options Options
}

// NewGitLabProvider creates a Provider given a token.
// If accessing private repositories, token must not be empty.
func NewGitLabProvider(token string, options Options) Provider {
	GitLabBaseURL := DefaultGitLabAPIURL
	if options.BaseURL != "" {
		GitLabBaseURL = options.BaseURL
	}

	client, err := gitlab.NewClient(token, gitlab.WithBaseURL(GitLabBaseURL))
	if err != nil {
		panic(fmt.Sprintf("could not set base URL for gitlab client: %v", err))
	}

	return &GitLabProvider{
		token:   token,
		client:  client,
		options: options,
	}
}

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
		private:     !r.Public,
	}
}

func (p *GitLabProvider) gatherAccessiblePage(page int, verbose bool) ([]GitRepository, int, error) {
	opt := &gitlab.ListProjectsOptions{
		ListOptions: gitlab.ListOptions{
			PerPage: reposPerPage, Page: page,
		}, Statistics: gitlab.Bool(true),
		Membership: gitlab.Bool(true),
	}

	if verbose {
		log.Infof("Gathering page %v for %v\n", page, p.client.BaseURL())
	}

	repos, resp, err := p.client.Projects.ListProjects(opt)
	if err != nil {
		return nil, 0, err
	}

	repositories := make([]GitRepository, 0, len(repos))

	for _, repo := range repos {
		if !p.options.IncludeForkedRepos && repo.ForkedFromProject != nil {
			continue
		}

		repositories = append(repositories, createFromGitlabRepo(repo))
	}

	return repositories, resp.TotalPages, nil
}

func (p *GitLabProvider) gatherGroupProjectPage(groupID, page int, verbose bool) ([]GitRepository, int, error) {
	opt := &gitlab.ListGroupProjectsOptions{
		ListOptions: gitlab.ListOptions{
			PerPage: reposPerPage,
			Page:    page,
		},
	}

	if verbose {
		log.Infof("Gathering page %v for %v\n", page, p.client.BaseURL())
	}

	repos, resp, err := p.client.Groups.ListGroupProjects(groupID, opt)
	if err != nil {
		return nil, 0, err
	}

	repositories := make([]GitRepository, 0, len(repos))

	for _, repo := range repos {
		if !p.options.IncludeForkedRepos && repo.ForkedFromProject != nil {
			continue
		}

		repositories = append(repositories, createFromGitlabRepo(repo))
	}

	return repositories, resp.TotalPages, nil
}

func (p *GitLabProvider) findGroup(name string) (int, error) {
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
		if strings.EqualFold(group.FullName, name) || strings.EqualFold(group.FullPath, name) {
			return group.ID, nil
		}
	}

	return 0, ErrGroupNotFound
}

// Gather gathers user's repositories for the configured token.
func (p *GitLabProvider) Gather(object string) ([]GitRepository, error) {
	// repositories protected by mu, since multiple goroutines will access it
	repositories := make([]GitRepository, 0)
	if object != "" {
		repositories = p.collectFromGroup(repositories, object)
	} else {
		repositories = p.collectAllAccessible(repositories)
	}

	return repositories, nil
}

func (p *GitLabProvider) collectAllAccessible(
	repositories []GitRepository) []GitRepository {
	wg := sync.WaitGroup{}

	var mu sync.Mutex

	_, totalPages, err := p.gatherAccessiblePage(1, false)
	if err != nil {
		log.Errorf("Error gathering first page: %v\n", err)

		return repositories
	}

	for pageCount := 1; pageCount <= totalPages; pageCount++ {
		wg.Add(1)

		go func(page int) {
			defer wg.Done()

			pageRepositories, _, err := p.gatherAccessiblePage(page, true)
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

func (p *GitLabProvider) collectFromGroup(repositories []GitRepository,
	object string) []GitRepository {
	groupID, err := p.findGroup(object)
	if err != nil {
		log.Errorf("Error finding group '%v' : %v\n", object, err)

		return repositories
	}

	wg := sync.WaitGroup{}

	var mu sync.Mutex

	_, totalPages, err := p.gatherGroupProjectPage(groupID, 1, false)
	if err != nil {
		log.Errorf("Error gathering first page: %v\n", err)

		return repositories
	}

	log.Infof("Gathering repositories for group %s\n", object)

	for pageCount := 1; pageCount <= totalPages; pageCount++ {
		wg.Add(1)

		go func(page int) {
			defer wg.Done()

			pageRepositories, _, err := p.gatherGroupProjectPage(groupID, page, true)
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
func (p *GitLabProvider) CloneRepository(
	ctx context.Context,
	cloner cloner.Cloner,
	repository GitRepository) (string, error) {
	url := repository.GetSSHUrl()
	// If token doesn't exist or if SSH cloning was specified, don't try to basic auth
	if p.token != "" && !p.options.SSHCloning {
		url = repository.GetHTTPUrl()
		url = strings.Replace(url, "https://", fmt.Sprintf("https://%s:%s@", p.token, p.token), 1)
	}

	return cloner.CloneRepository(ctx, url)
}
