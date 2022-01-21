package srcfingerprint

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"regexp"

	log "github.com/sirupsen/logrus"
)

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

// FastExtractor will directly extract the information without using an Analyzer
// There are designed to use raw git commands to get what is needed.
type FastExtractor struct {
	ChanGitFiles chan *GitFile
}

func (fe *FastExtractor) Run(path string, after string) chan *GitFile {
	log.Infof("Extracting commits from path %s\n", path)

	cmdRevList := "git rev-list --objects --all"

	if after != "" {
		cmdRevList = fmt.Sprintf("git rev-list --objects --all --after '%s'", after)
	}

	cmdBase := cmdRevList + "| git cat-file --batch-check='{\"sha\": \"%(objectname)\", \"type\": \"%(objecttype)\", \"filepath\": \"%(rest)\", \"size\": \"%(objectsize)\"}' | grep '\"type\": \"blob\"'" //nolint
	cmd := exec.Command("bash", "-c", cmdBase)
	cmd.Dir = path

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatalln(err)
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
				log.Infoln("finished reading all files from stdout from git")

				break
			}

			num++

			log.Debugf("parsing line %s", line)

			var gitFile GitFile

			// Replace backslashes by escaped backslashes
			re := regexp.MustCompile(`\\\\(.)`)
			cleanedLine := re.ReplaceAll(line, []byte("\\\\$1"))

			err := json.Unmarshal(cleanedLine, &gitFile)
			if err != nil {
				log.Warnln("Error while parsing", string(line), err)
			}

			fe.ChanGitFiles <- &gitFile
		}

		log.Infof("finished iterating over files, we have collected %d files\n", num)

		if err := os.RemoveAll(path); err != nil {
			log.Errorln("Unable to remove directory ", path)
		}

		log.Infof("Correctly removed cloned directory %s", path)
		close(fe.ChanGitFiles)
	}()

	return fe.ChanGitFiles
}
