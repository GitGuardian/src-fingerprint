# src-fingerprint

- [src-fingerprint](#src-fingerprint)
  - [Introduction](#introduction)
  - [Install](#install)
  - [Generate My Token](#generate-my-token)
    - [GitHub](#github)
    - [GitLab](#gitlab)
  - [Compute my fileshas](#compute-my-fileshas)
    - [GitHub](#github-1)
    - [GitLab](#gitlab-1)
    - [Bitbucket server (formely Atlassian Stash)](#bitbucket-server-formely-atlassian-stash)
    - [Repository](#repository)

## Introduction

The purpose of this package is to extract some git related information (all files sha, also commits sha) from your hosted source version control system

It supports 3 main on premise version control service:

- GitHub Enterprise
- Gitlab CE and EE
- Bitbucket

## Install

### Pre-compiled executables

Get them [here](http://github.com/gitguardian/src-fingerprint/releases).

### Source

You need `go` installed and `GOBIN` in your `PATH`. Once that is done, run the
command:

```shell
$ go get -u github.com/gitguardian/src-fingerprint
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

## Compute my fileshas

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
