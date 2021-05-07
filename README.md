# src-fingerprint

## Introduction

The purpose of this package is to extract some git related information (all files sha, also commits sha) from your hosted source version control system

It supports 3 main on premise version control service:

- GitHub Enterprise
- Gitlab CE and EE
- Bitbucket

## Providers

### GitHub

1. Export all file SHAs from a GitHub Org with private repositories to a file with logs:

```sh
env VCS_TOKEN="<token>" src-fingerprint -v --output file_shas_collected_dna.json --provider github --object GitGuardian
```

2. Export all file SHAs of every repository the user can access to `stdout`:

```sh
env VCS_TOKEN="<token>" src-fingerprint -v --provider github GitGuardian
```

### GitLab

1. Export all file SHAs from a GitLab group with private projects to a file with logs:

```sh
env VCS_TOKEN="<token>" src-fingerprint -v --output file_shas_collected_dna.json --provider gitlab --object "GitGuardian-dev-group"
```

2. Export all file SHAs of every project the user can access to `stdout`:

> :warning: On `gitlab.com` this will attempt to retrieve all repositories on `gitlab.com`

```sh
env VCS_TOKEN="<token>" src-fingerprint -v --provider gitlab
```

### Bitbucket server (formely Atlassian Stash)

1. Export all file SHAs from a Bitbucket project with private repository to a file with logs:

```sh
env VCS_TOKEN="<token>" src-fingerprint -v --output file_shas_collected_dna.json --provider bitbucket --object "GitGuardian Project"
```

2. Export all file SHAs of every repository the user can access to `stdout`:

```sh
env VCS_TOKEN="<token>" src-fingerprint -v --provider bitbucket
```

### Repository

Allows the processing of a single repository given a git clone URL

1. ssh cloning

```sh
src-fingerprint -p repository -o 'git@github.com:GitGuardian/gg-shield.git'
```

2. http cloning with basic authentication

```sh
src-fingerprint -p repository -o 'https://user:password@github.com/GitGuardian/gg-shield.git'
```

2. http cloning without basic authentication

```sh
src-fingerprint -p repository -o 'https://github.com/GitGuardian/gg-shield.git'
```

## Architecture

### Main overview

All the git information can be found inside commit that are located inside git repositories
Our tree element step are the following:

- Collect all repositories URL from the company.
- Clone them with the appropriate authentication.
- Run git commands to extract the information we need on each repository.
- Gather data and store this information in a json file.

### Implementation

The root package is the abstract implementation of the extractor. It contains a Cloner, that clones a git repository.
It contains a Pipeline that extracts git information for every git artifact (currently a git file but we could support commit), of every repository of an organization.

The github package contains the implementation of the Github Provider.
The gitlab package contains the implementation of the Gitlab Provider.

The cmd/src-fingerprint package contains the binary code. It reads from CLI and environment the configuration and run the Pipeline on an organization.

## Development build and testing

- Build binary

  ```sh
  go build ./cmd/src-fingerprint
  ```

- Set env var `VCS_TOKEN` to the GitHub Token or GitLab Token

  ```sh
  export VCS_TOKEN="<token>"
  ```

- Run and read doc

  ```sh
  ./src-fingerprint
  ```

- Run on a given user/group
  ```sh
  ./src-fingerprint --provider github --object Uber
  ./src-fingerprint --provider-url http://gitlab.example.com --provider gitlab --object Groupe
  ```

### Libraries we use

#### Providers

- GitHub wrapper: "github.com/google/go-github/v18/github"
- Gitlab go wrapper: "github.com/xanzy/go-gitlab"
- Bitbucket wrapper: "github.com/suhaibmujahid/go-bitbucket-server/bitbucket"
- Repository: None

#### Cloning

- go-git: https://github.com/src-d/go-git

### Issues

- Repo size seems not to work on go gitlab wrapper.
- Channels are cheap. Complex design overloading semantics isn't.
