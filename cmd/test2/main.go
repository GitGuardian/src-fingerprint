package main

import (
	"dnacollector"

	"github.com/caarlos0/env"
	log "github.com/sirupsen/logrus"
	git2 "gopkg.in/src-d/go-git.v4"
)

type config struct {
	GithubToken string `env:"GITHUB_TOKEN"`
	GitlabToken string `env:"GITLAB_TOKEN"`
}

func main() {
	conf := config{}
	if err := env.Parse(&conf); err != nil {
		log.Fatalf("Could not parse env: %v\n", err)
	}

	// Config log
	log.SetFormatter(&log.TextFormatter{
		DisableColors:   true,
		FullTimestamp:   true,
		TimestampFormat: "Jan _2 15:04:05.000000000",
	})
	log.SetReportCaller(true)
	log.SetLevel(log.InfoLevel)
	//var cloner dnacollector.Cloner = &dnacollector.DiskCloner{}
	//auth := &http.BasicAuth{
	//	Username: "ericfourrier",
	//	Password: conf.GithubToken,
	//}

	//repository, err := cloner.CloneRepository("https://github.com/uber/cadence.git", auth)
	//if err != nil {
	//	log.Panic(err)
	//
	//}
	repository, err := git2.PlainOpen("/Users/ericfourrier/Documents/GGCode/dna-collector/testdata/cadence")
	if err != nil {
		log.Panic(err)
	}

	extractor := dnacollector.NewFastExtractor()
	extractor.Run(repository)

	arrGitFiles := make([]*dnacollector.GitFile, 0)
	for gitFile := range extractor.ChanGitFiles {
		arrGitFiles = append(arrGitFiles, gitFile)
		log.Debug(gitFile)
	}

	log.Infof("length of files collected %d", len(arrGitFiles))
}
