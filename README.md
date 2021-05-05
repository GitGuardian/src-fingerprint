# src-fingerprint

## Introduction

The purpose of this package is to extract some git related information (all files sha, also commits sha) from your hosted source version control system

It supports 3 main on premise version control service:

- GitHub Enterprise
- Gitlab CE and EE
- Bitbucket

## Use the package

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

## Some examples

- Don't forget to build the package: `go build ./cmd/src-fingerprint`

1. Export all files sha from a GitHub Org to a file with logs: `./src-fingerprint -v --output file_shas_collected_dna.json --provider github GitGuardian`

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

The cmd/dna-collector package contains the binary code. It reads from CLI and environment the configuration and run the Pipeline on an organization.

### Library we use

#### Providers

- GitHub: "github.com/google/go-github/v18/github"
- Gitlab go wrapper: "github.com/xanzy/go-gitlab"
- bitbucket not supported yet

#### Cloning

- go-git: https://github.com/src-d/go-git

### Issues

- Repo size seems not to work on go gitlab wrapper.
- Channels are cheap. Complex design overloading semantics isn't.
