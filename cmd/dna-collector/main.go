package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"dnacollector"
	"dnacollector/github"
	"dnacollector/gitlab"
	"io"
	"os"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/caarlos0/env"
)

type config struct {
	GithubToken string `env:"GITHUB_TOKEN"`
	GitlabToken string `env:"GITLAB_TOKEN"`
}

func runExtract(pipeline *dnacollector.Pipeline, user string) chan dnacollector.PipelineEvent {
	// buffer it a bit so it won't block if this is going too fast
	ch := make(chan dnacollector.PipelineEvent, 100)

	go func(eventChannel chan dnacollector.PipelineEvent) {
		defer close(eventChannel)
		pipeline.ExtractRepositories(user, eventChannel)
	}(ch)

	return ch
}

type authorInfo struct {
	Name           string
	Email          string
	Count          int
	LastCommitDate time.Time
}

func main() {
	var (
		verbose        = flag.Bool("verbose", false, "set to add verbose logging")
		extractForks   = flag.Bool("extract-forks", false, "set to extract fork repositories when possible")
		inMemory       = flag.Bool("in-memory", false, "set to clone git repositories in memory. If not set, repositories are cloned in a temporary folder")
		outputFilename = flag.String("output", "-", "set to change output. \"-\" means standard output")
		prettyPrint    = flag.Bool("pretty", false, "set to pretty print to output file")
		clonersCount   = flag.Int("cloners", 10, "set to change the number of cloners. More cloners means more memory usage")
		providerURL    = flag.String("provider-url", "", "Base URL of the Git provider API. If not set, defaults URL are used.")
	)
	conf := config{}

	if err := env.Parse(&conf); err != nil {
		log.Fatalf("Could not parse env: %v\n", err)
	}

	flag.Parse()

	if flag.NArg() < 1 {
		fmt.Fprintf(os.Stderr, "%v [flags] provider user\n", os.Args[0])
		fmt.Fprintln(os.Stderr, "List of supported flags:")
		flag.PrintDefaults()
		os.Exit(1)
	}

	if *verbose {
		log.SetLevel(log.InfoLevel)
	} else {
		log.SetLevel(log.ErrorLevel)
	}

	output := os.Stdout
	if *outputFilename != "-" {
		changedOutput, err := os.Create(*outputFilename)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Could not open output file: %v\n", err)
			os.Exit(1)
		}
		output = changedOutput
		defer output.Close()
	}

	var cloner dnacollector.Cloner = &dnacollector.DiskCloner{}
	if *inMemory {
		cloner = &dnacollector.MemoryCloner{}
	}

	providerStr := flag.Arg(0)
	user := flag.Arg(1)

	providerOptions := dnacollector.ProviderOptions{
		OmitForks: !*extractForks,
		BaseURL:   *providerURL,
	}
	var provider dnacollector.Provider
	switch providerStr {
	case "github":
		provider = github.NewProvider(conf.GithubToken, providerOptions)
	case "gitlab":
		provider = gitlab.NewProvider(conf.GitlabToken, providerOptions)
	default:
		log.Fatalf("unknown provider: %v\n", provider)
	}

	pipeline := dnacollector.Pipeline{
		Provider: provider,
		Cloner:   cloner,
		Analyzer: &dnacollector.Analyzer{},

		ClonersCount: *clonersCount,
	}

	ticker := time.Tick(1 * time.Second)

	eventChannel := runExtract(&pipeline, user)

	// runtime stats
	var (
		totalRepo    int
		doneRepo     int
		authors      map[string]*authorInfo
		commitsCount int
	)

	authors = make(map[string]*authorInfo)
loop:
	for {
		select {
		case event, opened := <-eventChannel:
			if !opened {
				break loop
			}

			switch typedEvent := event.(type) {
			case dnacollector.RepositoryListPipelineEvent:
				totalRepo = len(typedEvent.Repositories)
			case dnacollector.RepositoryPipelineEvent:
				if typedEvent.Finished {
					doneRepo++
				}
			case dnacollector.ResultPipelineEvent:
				commitsCount++

				identity := typedEvent.Author.Name + typedEvent.Author.Email
				if _, identityExists := authors[identity]; !identityExists {
					authors[identity] = &authorInfo{}
				}
				commit := typedEvent.Commit
				authors[identity].Count++
				authors[identity].Name = typedEvent.Author.Name
				authors[identity].Email = typedEvent.Author.Email
				if commit.Author.When.UTC().After(authors[identity].LastCommitDate) {
					authors[identity].LastCommitDate = commit.Author.When.UTC()
				}
			}
		case <-ticker:
			if totalRepo == 0 {
				continue
			}

			log.Infof("%v/%v repos: ", doneRepo, totalRepo)
			log.Infof("%v distinct authors, %v commit analyzed\n", len(authors), commitsCount)
		}
	}

	log.Infoln("Final stats:")
	log.Infof("%v/%v repos: ", doneRepo, totalRepo)
	log.Infof("%v distinct authors, %v commit analyzed\n", len(authors), commitsCount)

	log.Infof("Dumping to output %v", *outputFilename)

	authorsList := make([]*authorInfo, 0, len(authors))
	for _, author := range authors {
		authorsList = append(authorsList, author)
	}

	var (
		jsonBytes []byte
		err       error
	)
	if *prettyPrint {
		jsonBytes, err = json.MarshalIndent(authorsList, "", "\t")
	} else {
		jsonBytes, err = json.Marshal(authorsList)
	}

	if err != nil {
		log.Fatalf("Could not marshal result to JSON: %v\n", err)
	}

	if _, err = io.Copy(output, bytes.NewReader(jsonBytes)); err != nil {
		log.Fatalf("Could not save output: %v\n", err)
	}
	log.Infof("Done")
}
