package cloner

import (
	"bytes"
	"context"
	"errors"
	"os"
	"os/exec"
	"strings"

	"github.com/shirou/gopsutil/v3/process"

	log "github.com/sirupsen/logrus"
)

func terminateProcessAndChildren(pid int32) error {
	var (
		proc     *process.Process
		children []*process.Process
		err      error
	)

	proc, err = process.NewProcess(pid)
	if err != nil {
		return err
	}

	children, err = proc.Children()
	if err != nil {
		log.WithError(err).Warnf("could not get children of pid#%v", pid)

		children = []*process.Process{}
	}

	err = proc.Terminate()
	if err != nil {
		return err
	}

	for _, child := range children {
		err = terminateProcessAndChildren(child.Pid)
		if err != nil {
			log.WithError(err).Warnf("could not terminate pid#%v", pid)
		}
	}

	return nil
}

func cloneGitRepository(ctx context.Context, destDir, gitRepoURL string) error {
	var outbuf, errbuf bytes.Buffer
	// git clone github.com/author/name.git /tmp/workdir/author-name/clone
	cmd := exec.Command("git", "clone", gitRepoURL, destDir)
	cmd.Stdout = &outbuf
	cmd.Stderr = &errbuf

	cmd.Env = append(os.Environ(), "GIT_TERMINAL_PROMPT=0")

	go func() {
		<-ctx.Done()

		if cmd.Process != nil && (cmd.ProcessState == nil || !cmd.ProcessState.Exited()) {
			if err := terminateProcessAndChildren(int32(cmd.Process.Pid)); err != nil {
				log.WithError(err).Error("Could not terminate git")
			}
		}
	}()

	err := cmd.Run()
	if exitError, ok := err.(*exec.ExitError); ok {
		stderr := strings.TrimSpace(errbuf.String())

		if exitError.ExitCode() == gitExitUnclean {
			log.WithError(err).WithFields(log.Fields{
				"op":     "gitError",
				"stderr": stderr,
			}).Warnf("missing repo")
		} else if ctx.Err() != nil {
			log.Errorf("timeout reached while cloning the repository")
		} else {
			log.WithError(err).WithFields(log.Fields{
				"op":     "gitError",
				"stderr": stderr,
			}).Errorf("unhandled git error")
		}

		return errors.New("")
	}

	return err
}
