# src-fingerprint

- [src-fingerprint](#src-fingerprint)
  - [Introduction](#introduction)
  - [Install](#install)
  - [Generate My Token](#generate-my-token)
    - [GitHub](#github)
    - [GitLab](#gitlab)
  - [Compute my code fingerprints](#compute-my-code-fingerprints)
    - [GitHub](#github-1)
    - [GitLab](#gitlab-1)
    - [Bitbucket server (formely Atlassian Stash)](#bitbucket-server-formely-atlassian-stash)
    - [Repository](#repository)
  - [License](#License)

## Introduction

The purpose of `src-fingerprint` is to provide an easy way to extract git related information (namely all file shas of a repository) from your hosted source version control system.

This util supports 3 main version control systems:

- GitHub and GitHub Enterprise
- Gitlab CE and EE
- Bitbucket

## Install

### Pre-compiled executables

Get the executables [here](http://github.com/gitguardian/src-fingerprint/releases).

### Using Homebrew

If you're using [Homebrew](https://brew.sh/index_fr) you can add GitGuardian's tap and then install src-fingerprint. Just run the following commands :

```shell
brew tap gitguardian/tap
brew install src-fingerprint
```

### From the sources

You need `go` installed and `GOBIN` in your `PATH`. Once that is done, run the
command:

```shell
$ go get -u github.com/gitguardian/src-fingerprint/cmd/src-fingerprint
```

## Generate My Token

### GitHub

1. Click on your profile picture at the top right of the screen. A dropdown menu will appear and you will be able to access your personal settings by clicking on _Settings_
2. On your profile, go to Developer Settings
3. Select Personal Access Tokens
4. Click on `Generate a new token`
5. Click the `repo` box. This is the only scope we need
6. Click on `Generate token`. The token will only be available at this time so make sure you keep it in a safe place

### GitLab

1. Click on your profile picture at the top right of the screen. A dropdown menu will appear and you will be able to access your personal settings by clicking on _Preferences_
2. In the left sidebar, click on `Access Tokens`
3. Click the `read repository` box. This is the only scope we need. You can set an end-date for the token validity if you want more security
4. Click on `Create personal token`. The token will only be available at this time so make sure you keep it in a safe place

## Compute my code fingerprints

### General information

The output format can be chosen between `jsonl`, `json`, `gzip-jsonl` and `gzip-json` with the option `--export-format`.  
The default format is `gzip-jsonl` to minimize the size of the output file.  
The default output filepath is `./fingerprints.jsonl.gz`. Use `--output` to override this behavior.  
Also, note that if you were to download fingerprints for repositories of a big organization, `src-fingerprint` has a limit to process no more than 100
repositories. You can override this limit with the option `--limit`, a limit of 0 will process all repos of the organization.

### Default behavior

Note that by default **for github provider**, `src-fingerprint` will exclude private repositories, forks and archived repositories from the fingerprints computation. Use options `-e` or `--all` to change this behavior.

### GitHub

1. Export all fingerprints from private repositories from a GitHub Org to the default path `./fingerprints.jsonl.gz` with logs:

```sh
env VCS_TOKEN="<token>" src-fingerprint -v --provider github --object ORG_NAME --all
```

2. Export all fingerprints of every repository the user can access to the default path `./fingerprints.jsonl.gz`:

```sh
env VCS_TOKEN="<token>" src-fingerprint -v --provider github --all
```

### GitLab

1. Export all fingerprints from private repositories of a GitLab group to the default path `./fingerprints.jsonl.gz` with logs:

```sh
env VCS_TOKEN="<token>" src-fingerprint -v --provider gitlab --object "GitGuardian-dev-group"
```

2. Export all fingerprints of every project the user can access to the default path `./fingerprints.jsonl.gz` with logs:

> :warning: On `gitlab.com` this will attempt to retrieve all repositories on `gitlab.com`

```sh
env VCS_TOKEN="<token>" src-fingerprint -v --provider gitlab
```

### Bitbucket server (formely Atlassian Stash)

1. Export all fingerprints from a Bitbucket project with private repository to the default path `./fingerprints.jsonl.gz` with logs:

```sh
env VCS_TOKEN="<token>" src-fingerprint -v --provider bitbucket --object "GitGuardian Project"
```

2. Export all fingerprints of every repository the user can access to the default path `./fingerprints.jsonl.gz` with logs:

```sh
env VCS_TOKEN="<token>" src-fingerprint -v --provider bitbucket
```

### Repository

Allows the processing of a single repository given a git clone URL

1. ssh cloning

```sh
src-fingerprint -p repository -u 'git@github.com:GitGuardian/gg-shield.git'
```

2. http cloning with basic authentication

```sh
src-fingerprint -p repository -u 'https://user:password@github.com/GitGuardian/gg-shield.git'
```

2. http cloning without basic authentication

```sh
src-fingerprint -p repository -u 'https://github.com/GitGuardian/gg-shield.git'
```

3. repository in a local directory

```sh
src-fingerprint -p repository -u /projects/gitlab/src-fingerprint
```

4. repository in current directory

```sh
src-fingerprint -p repository -u .
```

## License

GitGuardian `src-fingerprint` is MIT licensed.
