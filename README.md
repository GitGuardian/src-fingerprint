# git-users-extractor

## Introduction

The purpose of this package is to extract all git user information of your developers from your hosted source version control system

It supports 3 main on premise version control service:

* GitHub Enterprise
* Gitlab CE and EE
* BitBucket (not supported yet)

## Use the package

* Build binary
    ```sh
    go build ./cmd/dna-collector
    ```
* Set env var `GITHUB_TOKEN` or `GITLAB_TOKEN`
    ```sh
    export GITHUB_TOKEN="<token>"
    export GITLAB_TOKEN="<token>"
    ```
* Run and read doc
    ```sh
    ./dna-collector -help
    ```
* Run on a given user/group
    ```sh
    ./dna-collector github Uber
    ./dna-collector -provider-url http://gitlab.example.com gitlab Groupe
    ```

## Some examples 

* Don't forget to build the package:  `go build ./cmd/dna-collector`

1. Export all files sha from a GitHub Org to a file with logs: `./dna-collector -verbose -output file_shas_collected_dna.json github GitGuardian`
2. Name params should be passed before positional parameters.
## Architecture

### Main overview
All the git information can be found inside commit that are located inside git repositories
Our tree element step are the following:
* Collect all repositories URL from the company.
* Clone them with the appropriate authentication.
* Clone them and iterate over commits to extract git config information.
* Store this information in a json file.

### Implementation
The root package is the abstract implementation of the extractor. It contains a Cloner, that clones a git repository.
It contains a Pipeline that extracts git information for every commit, of every repository of an organization.

The github package contains the implementation of the Github Provider.
The gitlab package contains the implementation of the Gitlab Provider.

The cmd/guser-extractor package contains the binary code. It reads from CLI and environment the configuration and run the Pipeline on an organization.

### Library we use

#### Providers
* GitHub: "github.com/google/go-github/v18/github"
* Gitlab go wrapper: "github.com/xanzy/go-gitlab"
* bitbucket not supported yet

#### Cloning
* go-git: https://github.com/src-d/go-git


### Issues
* Repo size seems not to work on go gitlab wrapper.
* Channels are cheap. Complex design overloading semantics isn't.


### Notes eric 

* We want to add more extractors -> We need to extract file shas
* We will add more analyzer -> From the commits we want to get commits shas
