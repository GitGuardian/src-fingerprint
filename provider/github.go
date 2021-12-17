package provider

import (
	"context"
	"fmt"
	"net/url"
	"srcfingerprint/cloner"
	"strings"
	"sync"

	"github.com/google/go-github/v36/github"
	log "github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
)

const (
	// DefaultGithubAPIURL is the default API URL.
	DefaultGithubAPIURL = "https://api.github.com/"
	unknownTotal        = 99999
)

// GitHubProvider is capable of gathering Github repositories from an org.
type GitHubProvider struct {
	client     *github.Client
	options    Options
	token      string
	totalPages int
	isOrg      bool
}

func createFromGithubRepo(r *github.Repository) *Repository {
	return &Repository{
		name:        r.GetName(),
		sshURL:      r.GetSSHURL(),
		httpURL:     r.GetHTMLURL(),
		createdAt:   r.GetCreatedAt().Time,
		storageSize: int64(r.GetSize()),
		private:     r.GetPrivate(),
	}
}

// NewGitHubProvider creates a new Github Provider.
func NewGitHubProvider(token string, options Options) Provider {
	client := github.NewClient(oauth2.NewClient(
		context.Background(),
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
		client:     client,
		options:    options,
		token:      token,
		totalPages: unknownTotal,
		isOrg:      true,
	}
}

// Gather Page for GitHub provider.
// If is first page update the total page count and try as user as well.
func (p *GitHubProvider) gatherPage(user string, page int) ([]GitRepository, error) {
	total := fmt.Sprint(p.totalPages)
	if total == fmt.Sprint(unknownTotal) {
		total = "?"
	}

	log.Infof("Gathering page %v/%v for %s\n", page, total, user)

	var (
		resp       *github.Response
		repos      []*github.Repository
		collectErr error
		visibility string
	)

	if p.options.AllRepositories {
		visibility = "all"
	} else {
		visibility = "private"
	}

	if p.isOrg {
		opt := &github.RepositoryListByOrgOptions{
			ListOptions: github.ListOptions{
				PerPage: reposPerPage, Page: page,
			},
			Type: visibility,
		}
		repos, resp, collectErr = p.client.Repositories.ListByOrg(context.Background(), user, opt)

		if resp.StatusCode == 404 && page == 1 {
			p.isOrg = false
		}
	}

	if !p.isOrg {
		opt := &github.RepositoryListOptions{
			ListOptions: github.ListOptions{
				PerPage: reposPerPage, Page: page,
			},
			Visibility: visibility,
		}

		repos, resp, collectErr = p.client.Repositories.List(context.Background(), user, opt)
	}

	if collectErr != nil {
		return nil, collectErr
	}

	if p.totalPages == unknownTotal {
		p.totalPages = resp.LastPage
	}

	repositories := make([]GitRepository, 0, len(repos))

	for _, repo := range repos {
		if p.options.OmitForks && repo.GetFork() {
			continue
		}

		if p.options.SkipArchived && repo.Archived != nil && *repo.Archived {
			continue
		}

		repositories = append(repositories, createFromGithubRepo(repo))
	}

	return repositories, nil
}

// Gather gather user's git repositories and send them to outputChannel.
func (p *GitHubProvider) Gather(user string) ([]GitRepository, error) {
	log.Debugf("Gathering repositories for Github org %s\n", user)

	wg := sync.WaitGroup{}

	var mu sync.Mutex

	// repositories protected by mu, since multiple goroutines will access it
	repositories := make([]GitRepository, 0)

	for pageCount := 1; pageCount <= p.totalPages; pageCount++ {
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

		if pageCount == 1 {
			wg.Wait()

			if p.totalPages == unknownTotal {
				return nil, fmt.Errorf("unable to gather total pages")
			}
		}
	}

	wg.Wait()

	return repositories, nil
}

// CloneRepository clones a Github repository given the token. The token must have the `read_repository` rights.
func (p *GitHubProvider) CloneRepository(cloner cloner.Cloner,
	repository GitRepository) (string, error) {
	url := repository.GetHTTPUrl()
	// If token doesn't exist, don't try to basic auth
	if p.token != "" {
		url = strings.Replace(url, "https://", fmt.Sprintf("https://x-access-token:%s@", p.token), 1)
	}

	return cloner.CloneRepository(url)
}
