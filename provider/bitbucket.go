package provider

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"srcfingerprint/cloner"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/suhaibmujahid/go-bitbucket-server/bitbucket"
)

// BitbucketProvider is capable of gathering Bitbucket server repositories from an org.
type BitbucketProvider struct {
	client    *bitbucket.Client
	transport *AuthHeaderTransport
	options   Options
	token     string
}

const LastPage = -1
const BitbucketTimeout = 5 * time.Second
const BitbucketClientTimeout = 10 * time.Second

func createFromBitbucketRepo(r *bitbucket.Repository) *Repository {
	sshURL := ""
	httpURL := ""

	for _, link := range r.Links.Clone {
		if link.Name == "ssh" {
			sshURL = link.Href
		} else if link.Name == "http" || link.Name == "https" {
			httpURL = link.Href
		}
	}

	httpURL = strings.Replace(httpURL, "http://", "https://", 1)

	return &Repository{
		name:        r.Name,
		sshURL:      sshURL,
		httpURL:     httpURL,
		createdAt:   time.Now(),
		storageSize: 0,
	}
}

type AuthHeaderTransport struct {
	T      http.RoundTripper
	token  string
	user   string
	userID string
}

func (at *AuthHeaderTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Add("Authorization", "Bearer "+at.token)
	resp, err := at.T.RoundTrip(req)

	if resp != nil && resp.Header.Get("x-auserid") != "" {
		at.userID = resp.Header.Get("x-auserid")
		at.user = resp.Header.Get("x-ausername")
	}

	return resp, err
}

func NewAuthHeaderTransport(T http.RoundTripper, token string) *AuthHeaderTransport {
	if T == nil {
		T = &http.Transport{
			Dial: (&net.Dialer{
				Timeout: BitbucketTimeout,
			}).Dial,
			TLSHandshakeTimeout: BitbucketTimeout,
		}
	}

	return &AuthHeaderTransport{
		T:      T,
		token:  token,
		user:   "",
		userID: "",
	}
}

// NewBitbucketProvider creates a new Bitbucket provider.
func NewBitbucketProvider(token string, options Options) Provider {
	// BaseURL should be like http://localhost:7990/rest/api/1.0/
	if options.BaseURL == "" {
		panic("This provider requires an API URL of form: http://examplebb.com/rest/api/1.0/")
	}

	_, err := url.Parse(options.BaseURL)
	if err != nil {
		panic(fmt.Sprintf("Bitbucket Base URL is not a valid url: %v. Should be of form: http://examplebb.com/rest/api/1.0/", options.BaseURL)) // nolint
	}

	transport := NewAuthHeaderTransport(nil, token)
	netClient := &http.Client{
		Timeout:   BitbucketClientTimeout,
		Transport: transport,
	}

	client, err := bitbucket.NewServerClient(options.BaseURL, netClient)
	if err != nil {
		panic(fmt.Sprintf("Unable to create bitbucket client: %v.", err))
	}

	return &BitbucketProvider{client: client, options: options, token: token, transport: transport}
}

func (p *BitbucketProvider) gatherRepos(start int, project string) ([]GitRepository, int, error) {
	opt := &bitbucket.ListRepositoriesOptions{ListOptions: bitbucket.ListOptions{Start: start, Limit: reposPerPage}}
	if project != "" {
		opt.ProjectName = project
	}

	log.Infof("Gathering repos %v -> %v\n", start, start+reposPerPage)

	repos, resp, err := p.client.Repositories.List(context.Background(), opt)
	if err != nil {
		return nil, 0, err
	}

	repositories := make([]GitRepository, 0, len(repos))

	for _, repo := range repos {
		if repo.Origin != nil {
			continue
		}

		repositories = append(repositories, createFromBitbucketRepo(repo))
	}

	if resp.IsLastPage {
		return repositories, LastPage, nil
	}

	return repositories, resp.NextPageStart, nil
}

func (p *BitbucketProvider) collect(
	repositories []GitRepository, project string) []GitRepository {
	var start = 0
	for start != LastPage {
		var err error

		var pageRepositories []GitRepository

		pageRepositories, start, err = p.gatherRepos(start, project)
		if err != nil {
			log.Errorf("Error gathering start %v:%v\n", start, err)

			continue
		}

		repositories = append(repositories, pageRepositories...)
	}

	return repositories
}

// Gather gather user's git repositories and send them to outputChannel.
func (p *BitbucketProvider) Gather(user string) ([]GitRepository, error) {
	log.Infof("Gathering repositories for Bitbucket %s\n", user)

	repositories := make([]GitRepository, 0)

	repositories = p.collect(repositories, user)

	return repositories, nil
}

// CloneRepository clones a Github repository given the token. The token must have the `read_repository` rights.
func (p *BitbucketProvider) CloneRepository(cloner cloner.Cloner,
	repository GitRepository) (string, error) {
	authURL := repository.GetSSHUrl()
	// If token doesn't exist or if SSH cloning was specified, don't try to basic auth
	if p.token != "" && !p.options.SSHCloning {
		authURL = repository.GetHTTPUrl()
		authURL = strings.Replace(authURL,
			"https://", fmt.Sprintf("https://%s:%s@",
				p.transport.user, url.QueryEscape(p.token)), 1)
	}

	return cloner.CloneRepository(authURL)
}
