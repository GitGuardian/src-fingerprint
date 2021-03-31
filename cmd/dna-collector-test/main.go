package main

import (
	"dnacollector"
	"encoding/json"
	"errors"
	"io"

	"github.com/caarlos0/env"
	log "github.com/sirupsen/logrus"
	git2 "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing/format/diff"
	git "gopkg.in/src-d/go-git.v4/plumbing/object"
)

type config struct {
	GithubToken string `env:"GITHUB_TOKEN"`
	GitlabToken string `env:"GITLAB_TOKEN"`
}

// ErrFileSimplifiedCreation is the error returned when GitFileSimplified can not be instantiated.
var ErrFileSimplifiedCreation = errors.New("we could not instantiate GitFileSimplified from GitFile")

type GitFileSimplified struct {
	Name     string `json:"name"`
	Sha      string `json:"sha"`
	IsBinary bool   `json:"is_binary"`
	Size     int64  `json:"size"`
}

type CommitSimplified struct {
	Message   string               `json:"message"`
	Sha       string               `json:"sha"`
	Author    git.Signature        `json:"author"`
	Committer git.Signature        `json:"committer"`
	Files     []*GitFileSimplified `json:"files"`
}

func NewFromGitFile(file *git.File) *GitFileSimplified {
	isBinary, _ := file.IsBinary()

	return &GitFileSimplified{Name: file.Name, Sha: file.Hash.String(), IsBinary: isBinary, Size: file.Size}
}

func NewFromCommit(commit *git.Commit, files []*GitFileSimplified) *CommitSimplified {
	return &CommitSimplified{
		Message:   commit.Message,
		Sha:       commit.Hash.String(),
		Author:    commit.Author,
		Committer: commit.Committer,
		Files:     files,
	}
}

func NewFromFilePatch(filePatch diff.FilePatch) (*GitFileSimplified, error) {
	isBinary := filePatch.IsBinary()
	from, to := filePatch.Files()
	// If the patch creates a new file, "from" will be nil.
	// If the patch deletes a file, "to" will be nil.

	// Rare usecase
	switch {
	case to == nil && from == nil:
		return nil, ErrFileSimplifiedCreation
	case to != nil:
		// File creation
		return &GitFileSimplified{Name: to.Path(), Sha: to.Hash().String(), IsBinary: isBinary, Size: 0}, nil
	default:
		// File deletion
		return &GitFileSimplified{Name: from.Path(), Sha: from.Hash().String(), IsBinary: isBinary, Size: 0}, nil
	}
}

func NewAnalyzer() *Analyzer {
	return &Analyzer{make([]*CommitSimplified, 0)}
}

type Analyzer struct {
	CommitsList []*CommitSimplified
}

func (a *Analyzer) GetFilesFromCommit(commit *git.Commit) ([]*GitFileSimplified, error) {
	var files []*GitFileSimplified

	parent, err := commit.Parent(0)

	// There is no parent, so we take all the files
	if err != nil {
		filesIter, err := commit.Files()
		if err != nil {
			return nil, err
		}

		_ = filesIter.ForEach(func(file *git.File) error {
			fileSimplified := NewFromGitFile(file)
			log.Debugf("Appending file %s", fileSimplified.Name)
			if fileSimplified.Size > 0 {
				// fileSimplifiedJson, _ := json.Marshal(fileSimplified)
				log.Info(fileSimplified)
			}

			files = append(files, fileSimplified)

			return nil
		})
	} else {
		// There is a parent, so we consider only the diff
		patch, _ := commit.Patch(parent)
		filePatches := patch.FilePatches()
		// log.Info(patch.Stats())
		for _, fp := range filePatches {
			fileSimplified, err := NewFromFilePatch(fp)
			//for _, chunk := range fp.Chunks() {
			//	log.Debug(chunk)
			//}
			log.Debugf("Appending file %s", fileSimplified.Name)
			if err != nil {
				log.Warn(fileSimplified)
				files = append(files, fileSimplified)
			} else {
				continue
				// log.Error(ErrFileSimplifiedCreation)
				// log.Warn(commit)
				// log.Warn(fp)
			}
		}
	}

	return files, nil
}

// AnalyzeCommit extracts author and committer from a commit.
func (a *Analyzer) AnalyzeCommit(commit *git.Commit) string {
	// Store commmit sha
	files, _ := a.GetFilesFromCommit(commit)
	a.CommitsList = append(a.CommitsList, NewFromCommit(commit, files))

	return commit.Hash.String()
}

func (a *Analyzer) GetStats() map[string]int {
	res := make(map[string]int)
	res["nb_commits"] = len(a.CommitsList)
	nbFilesSHAs := 0

	for _, commit := range a.CommitsList {
		nbFilesSHAs += len(commit.Files)
	}

	res["nb_files_shas"] = nbFilesSHAs

	return res
}

//func (a *Analyzer) GetCommitShasArr() []string {
//	var res []string
//	for k := range a.SetCommitsSha {
//		res = append(res, k)
//	}
//	return res
//}

func main() {
	conf := config{}

	// Config log
	log.SetFormatter(&log.TextFormatter{
		DisableColors: true,
		FullTimestamp: true,
	})
	log.SetReportCaller(true)
	log.SetLevel(log.InfoLevel)

	if err := env.Parse(&conf); err != nil {
		log.Fatalf("Could not parse env: %v\n", err)
	}

	log.Debugln(conf)
	//var cloner dnacollector.Cloner = &dnacollector.MemoryCloner{}
	//auth := &http.BasicAuth{
	//	Username: "ericfourrier",
	//	Password: conf.GithubToken,
	//}

	repository, err := git2.PlainOpen("/Users/ericfourrier/Documents/GGCode/dna-collector/testdata/react-vis")
	if err != nil {
		log.Warnln(err)
	}

	_, err = repository.Config()
	// config is not used?
	if err != nil {
		log.Warnln(err)
	}

	// log.Infof("Cloned repo %v (size: %v)\n", repository.n, repository.GetStorageSize())
	extractor, err := dnacollector.NewExtractor(repository)
	if err != nil {
		log.Warnln(err)
	}

	analyzer := NewAnalyzer()

	for {
		commit, err := extractor.ExtractNextCommit()
		if err != nil && err != io.EOF {
			log.Panic(err)
		}

		if commit == nil {
			break
		}

		analyzer.AnalyzeCommit(commit)
	}

	res2, _ := json.Marshal(analyzer.CommitsList)
	log.Debug(string(res2))
	//fmt.Print(analyzer.SetCommitsSha)
	//for _, files := range analyzer.CommitTable {
	//	for _, file := range files {
	//		log.Info(file.Sha)
	//	}
	//}
	log.Info(analyzer.GetStats())
	log.Infof("Done extracting %v\n", repository)
}
