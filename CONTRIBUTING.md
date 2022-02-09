# Contributing

## Architecture

### Main overview

All the git information can be found inside commits that are located inside git repositories
Our tree element steps are the following:

- Collect all repository URL's from an object (org, user, group).
- Clone them with the appropriate authentication.
- Run git commands to extract the information we need on each repository.
- Gather data and store this information in a json file.

### Implementation

The root package is the abstract implementation of the extractor.

It contains a Pipeline that extracts git information for every git artifact
(currently a git file but we could support commit), of every repository of an organization.

The cmd/src-fingerprint package contains the binary code.
It reads from CLI and environment the configuration and run the Pipeline on an organization.

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
  ./src-fingerprint collect --provider github --object Uber
  ./src-fingerprint collect --provider-url http://gitlab.example.com --provider gitlab --object Groupe
  ```

## Performance considerations

Streaming is prefered in this scenario to avoid accumulation in memory of objects.

What we have done for now to improve performance:

- Write object by object to output/file by using jsonl format by default
- Clone using the native git executable. Natively written libraries tend to clone
  in memory at some point.

### To consider

- Limiting go channel numbers

## Libraries we use

#### Providers

- GitHub wrapper: "github.com/google/go-github/v36/github"
- Gitlab go wrapper: "github.com/xanzy/go-gitlab"
- Bitbucket wrapper: "github.com/suhaibmujahid/go-bitbucket-server/bitbucket"
- Repository: None

#### Cloning

- native wrapped git command

Using go-git resulted in in-memory cloning (stream to memory and then to directory).
This caused too high peaks of memory unsuitable for small VMs.

## Packaging

Packaging is done using [GoReleaser](https://goreleaser.com/) and
[nFPM](https://nfpm.goreleaser.com/).

You can test packaging using `make dist`.
