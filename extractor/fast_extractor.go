package extractor

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"

	log "github.com/sirupsen/logrus"
)

// FastExtractor will directly extract the information without using an Analyzer
// There are designed to use raw git commands to get what is needed.
type FastExtractor struct {
	ChanGitFiles chan *GitFile
}

type FastExtractorMaker struct{}

func (fem *FastExtractorMaker) Make() Extractor {
	return FastExtractor{make(chan *GitFile)}
}

func (fe FastExtractor) Next() (*GitFile, bool) {
	gitFile, ok := <-fe.ChanGitFiles

	return gitFile, ok
}

func (fe FastExtractor) Run(path string, after string) {
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

			err := json.Unmarshal(line, &gitFile)
			if err != nil {
				log.Warnln(err)
			}

			fe.ChanGitFiles <- &gitFile
		}

		close(fe.ChanGitFiles)
		log.Infof("finishing iterating over files, we have collected %d files\n", num)

		if err := os.RemoveAll(path); err != nil {
			log.Errorln("Unable to remove directory ", path)
		}
	}()
}
