# src-fingerprint

- [Introduction](#introduction)
- [Installation](#installation)
  - [Using pre-compiled executables](#using-pre-compiled-executables)
  - [Installing from sources](#installing-from-sources)
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

This util's main command is the `collect` command used to collect source code fingerprints from a version control system or a local repository. It supports 3 main VCS:

- GitHub and GitHub Enterprise
- Gitlab CE and EE
- Bitbucket

## Installation

### Using pre-compiled executables

#### macOS, using Homebrew

If you're using [Homebrew](https://brew.sh/index) you can add GitGuardian's tap and then install src-fingerprint. Just run the following commands:

```shell
brew tap gitguardian/tap
brew install src-fingerprint
```

#### Linux packages

Deb and RPM packages are available on [Cloudsmith](https://cloudsmith.io/~gitguardian/repos/src-fingerprint/packages/).

Setup instructions:

- [Deb packages](https://cloudsmith.io/~gitguardian/repos/src-fingerprint/setup/#formats-deb)
- [RPM packages](https://cloudsmith.io/~gitguardian/repos/src-fingerprint/setup/#formats-rpm)

#### Windows

Open a PowerShell prompt and run this command:

```shell
iwr -useb https://raw.githubusercontent.com/GitGuardian/src-fingerprint/main/scripts/windows-installer.ps1 | iex
```

The script asks for the installation directory. To install silently, use these commands instead:

```shell
iwr -useb https://raw.githubusercontent.com/GitGuardian/src-fingerprint/main/scripts/windows-installer.ps1 -Outfile install.ps1
.\install.ps1 C:\Destination\Dir
rm install.ps1
```

Note that `src-fingerprint` requires Unix commands such as `bash` to be available, so it runs better from a "Git Bash" prompt.

#### Manual download

You can also download the archives directly from the [releases page](http://github.com/gitguardian/src-fingerprint/releases).

### Installing from sources

You need `go` installed and `GOBIN` in your `PATH`. Once that is done, run the command:

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

## Collect my code fingerprints

### General information

The output format can be chosen between `jsonl`, `json`, `gzip-jsonl` and `gzip-json` with the option `--export-format`.  
The default format is `gzip-jsonl` to minimize the size of the output file.  
The default output filepath is `./fingerprints.jsonl.gz`. Use `--output` to override this behavior.  
Also, note that if you were to download fingerprints for repositories of a big organization, `src-fingerprint` has a limit to process no more than 100
repositories. You can override this limit with the option `--limit`, a limit of 0 will process all repos of the organization.

### Sample output

Here is an example of some lines of a `.jsonl` format output:

```shell
{"repository_name":"src-fingerprint","private":false,"sha":"a0c16efce5e767f04ba0c6988d121147099a17df","type":"blob","filepath":".env.example","size":"31"}
{"repository_name":"src-fingerprint","private":false,"sha":"d425eb0f8af66203dbeef50c921ea5bff0f2acba","type":"blob","filepath":".github/workflows/tag.yml","size":"882"}
{"repository_name":"src-fingerprint","private":false,"sha":"c7f341033d78474b125dd56d8adaa3f0fc47faf2","type":"blob","filepath":".github/workflows/test.yml","size":"899"}
{"repository_name":"src-fingerprint","private":false,"sha":"f4409d88950abd4585d8938571864726533a7fa5","type":"blob","filepath":".gitignore","size":"356"}
{"repository_name":"src-fingerprint","private":false,"sha":"f733f951ace2e032c270d2f3cf79c2efb8187b5b","type":"blob","filepath":".gitlab-ci.yml","size":"85"}
{"repository_name":"src-fingerprint","private":false,"sha":"d17ae66a017477bc65a2f433bf23d551ffc6bd75","type":"blob","filepath":".golangci.yml","size":"1196"}
{"repository_name":"src-fingerprint","private":false,"sha":"ee08a617cfb1c63c1c55fa4cb15e8bac0095346f","type":"blob","filepath":".goreleaser.yml","size":"2127"}
```

### Default behavior

Note that by default, `src-fingerprint` will exclude forked repositories from the fingerprints computation. **For GitHub provider** archived repositories and public repositories will also be excluded by default. Use flags `--include-forked-repos`, `--include-archived-repos` or `include-public-repos` to change this behavior.

### GitHub

1. Export all fingerprints from private repositories from a GitHub Org to the default path `./fingerprints.jsonl.gz` with logs:

```sh
env VCS_TOKEN="<token>" src-fingerprint -v --provider github --object ORG_NAME
```

2. Export all fingerprints of every repository the user can access to the default path `./fingerprints.jsonl.gz`:

```sh
env VCS_TOKEN="<token>" src-fingerprint -v --provider github --include-public-repos --include-forked-repos --include-archived-repos
```

### GitLab

1. Export all fingerprints from private repositories of a GitLab group to the default path `./fingerprints.jsonl.gz` with logs:

```sh
env VCS_TOKEN="<token>" src-fingerprint -v --provider gitlab --object "GitGuardian-dev-group"
```

2. Export all fingerprints of every project the user can access to the default path `./fingerprints.jsonl.gz` with logs:

```sh
env VCS_TOKEN="<token>" src-fingerprint -v --provider gitlab --include-forked-repos
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
src-fingerprint collect -p repository -u 'git@github.com:GitGuardian/gg-shield.git'
```

2. http cloning with basic authentication

```sh
src-fingerprint collect -p repository -u 'https://user:password@github.com/GitGuardian/gg-shield.git'
```

2. http cloning without basic authentication

```sh
src-fingerprint collect -p repository -u 'https://github.com/GitGuardian/gg-shield.git'
```

3. repository in a local directory

```sh
src-fingerprint collect -p repository -u /projects/gitlab/src-fingerprint
```

4. repository in current directory

```sh
src-fingerprint collect -p repository -u .
```

## License

GitGuardian `src-fingerprint` is MIT licensed.
