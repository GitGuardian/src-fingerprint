package dnacollector

import (
	"bufio"
	"encoding/json"
	"errors"
	"os"
	"os/exec"

	log "github.com/sirupsen/logrus"
	"gopkg.in/src-d/go-billy.v4/helper/chroot"
	git "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
	"gopkg.in/src-d/go-git.v4/storage/filesystem"
)

// Extractor extracts commits for a given repository.
type Extractor struct {
	iter object.CommitIter
}

// NewExtractor creates a new Extractor.
func NewExtractor(repository *git.Repository) (*Extractor, error) {
	iter, err := repository.CommitObjects()
	if err != nil {
		return nil, err
	}

	return &Extractor{
		iter: iter,
	}, nil
}

// ExtractNextCommit extract the next  commit of the repository.
// It returns an io.EOF error if there are no more commits.
func (e *Extractor) ExtractNextCommit() (*object.Commit, error) {
	return e.iter.Next()
}

// FastExtractor will directly extract the information without using an Analyzer
// There are designed to use raw git commands to get what is needed

func GetBasePathGoGitRepo(r *git.Repository) (string, error) {
	//
	// Try to grab the repository Storer
	s, ok := r.Storer.(*filesystem.Storage)
	if !ok {
		return "", errors.New("repository storage is not filesystem.Storage")
	}

	// Try to get the underlying billy.Filesystem
	fs, ok := s.Filesystem().(*chroot.ChrootHelper)
	if !ok {
		return "", errors.New("filesystem is not chroot.ChrootHelper")
	}

	return fs.Root(), nil
}

//type BaseGitFileIterator struct {
//	repository *git.Repository
//	filesChan  chan *GitFile
//}
//
//func (i *BaseGitFileIterator) Compute(repository *git.Repositor) error {
//
//}
//
//func (i *BaseGitFileIterator) Close() error {
//
//}

type BaseExtractor interface {
	Next() (interface{}, error)
}

type GitFile struct {
	Sha      string `json:"sha"`
	Type     string `json:"type"`
	Filepath string `json:"filepath"`
	Size     string `json:"size"`
}

func NewFastExtractor() *FastExtractor {
	return &FastExtractor{make(chan *GitFile)}
}

type FastExtractor struct {
	ChanGitFiles chan *GitFile
}

func (fe *FastExtractor) Run(repository *git.Repository) chan *GitFile {
	// https://gist.github.com/ochinchina/9e409a88e77c3cfd94c3
	path, err := GetBasePathGoGitRepo(repository)
	if err != nil {
		log.Fatal(err)
	}

	err = os.Chdir(path)
	if err != nil {
		log.Fatal(err)
	}

	log.Infof("Extracting commits from path %s", path)
	cmdBase := "git rev-list --objects --all | git cat-file --batch-check='{\"sha\": \"%(objectname)\", \"type\": \"%(objecttype)\", \"filepath\": \"%(rest)\", \"size\": \"%(objectsize:disk)\"}' | grep '\"type\": \"blob\"'" //nolint
	cmd := exec.Command("bash", "-c", cmdBase)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}

	err = cmd.Start()
	if err != nil {
		log.Fatal(err)
	}

	buf := bufio.NewReader(stdout) // Notice that this is not in a loop
	num := 0

	go func() {
		for {
			line, _, _ := buf.ReadLine()
			if len(line) == 0 {
				log.Info("finish reading all files from stdout from git")

				break
			}

			num++

			log.Debugf("parsing line %s", line)

			var gitFile GitFile

			err := json.Unmarshal(line, &gitFile)
			if err != nil {
				log.Warnln(err)
			}

			fe.ChanGitFiles <- &gitFile
		}

		close(fe.ChanGitFiles)
		log.Info("channel is closed")
		log.Infof("finishing iterating over files, we have collected %d files", num)
	}()

	return fe.ChanGitFiles
}

//targetFile := fe.BaseDirStorage + "/" + "sha_files.jsonl"
//log.Infof("Saving to file %s", targetFile)
//dataFile, err := os.Create(targetFile)
//if err != nil {
//	log.Error(err)
//}
//defer dataFile.Close()
//_, err = dataFile.WriteString(out.String())
//return err
