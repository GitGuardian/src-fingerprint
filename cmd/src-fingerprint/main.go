package main

import (
	"fmt"
	"io"
	"os"
	"srcfingerprint"
	"srcfingerprint/cloner"
	"srcfingerprint/exporter"
	"srcfingerprint/provider"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

var version = "unknown"
var builtBy = "unknown"
var date = "unknown"

const MaxPipelineEvents = 100

func runExtract(pipeline *srcfingerprint.Pipeline, user string, after string) chan srcfingerprint.PipelineEvent {
	// buffer it a bit so it won't block if this is going too fast
	ch := make(chan srcfingerprint.PipelineEvent, MaxPipelineEvents)

	go func(eventChannel chan srcfingerprint.PipelineEvent) {
		defer close(eventChannel)
		pipeline.ExtractRepositories(user, after, eventChannel)
	}(ch)

	return ch
}

func getProvider(providerStr string, token string, providerOptions provider.Options) (provider.Provider, error) {
	switch providerStr {
	case "github":
		return provider.NewGitHubProvider(token, providerOptions), nil
	case "gitlab":
		return provider.NewGitLabProvider(token, providerOptions), nil
	case "bitbucket":
		return provider.NewBitbucketProvider(token, providerOptions), nil
	case "repository":
		return provider.NewGenericProvider(providerOptions), nil
	default:
		return nil, fmt.Errorf("invalid provider string: %s", providerStr)
	}
}

func getExporter(exporterStr string, output io.WriteCloser) (exporter.Exporter, error) {
	switch exporterStr {
	case "json":
		return exporter.NewJSONExporter(output), nil
	case "gzip-json":
		return exporter.NewGzipJSONExporter(output), nil
	case "jsonl":
		return exporter.NewJSONLExporter(output), nil
	case "gzip-jsonl":
		return exporter.NewGzipJSONLExporter(output), nil
	default:
		return nil, fmt.Errorf("invalid export format: %s", exporterStr)
	}
}

type authorInfo struct {
	Name           string
	Email          string
	Count          int
	LastCommitDate time.Time
}

const DefaultClonerN = 8

func main() {
	cli.VersionFlag = &cli.BoolFlag{
		Name:  "version",
		Usage: "print version",
	}

	cli.VersionPrinter = func(c *cli.Context) {
		log.Printf("src-fingerprint version=%s date=%s builtBy=%s\n", version, date, builtBy)
	}

	app := &cli.App{
		Name:    "src-fingerprint",
		Version: version,
		Usage:   "Collect user/organization file hashes from your vcs provider of choice",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "verbose",
				Aliases: []string{"v"},
				Value:   false,
				Usage:   "verbose logging",
			},
			&cli.BoolFlag{
				Name:    "extract-forks",
				Aliases: []string{"e"},
				Value:   false,
				Usage:   "extract fork repositories when possible",
			},
			&cli.BoolFlag{
				Name:  "skip-archived",
				Value: false,
				Usage: "skip archived repositories",
			},
			&cli.StringFlag{
				Name:    "output",
				Aliases: []string{"o"},
				Value:   "-",
				Usage:   "set output path to `FILE`. stdout by default",
			},
			&cli.StringFlag{
				Name:  "export-format",
				Value: "jsonl",
				Usage: "export format: 'jsonl'/'gzip-jsonl'/'json'/'gzip-json'. 'jsonl' by default",
			},
			&cli.StringFlag{
				Name:  "clone-dir",
				Value: "-",
				Usage: "set cloning location for repositories",
			},
			&cli.StringFlag{
				Name:  "after",
				Value: "",
				Usage: "set a commit date after which we want to collect fileshas",
			},
			&cli.StringFlag{
				Name:     "provider",
				Aliases:  []string{"p"},
				Required: true,
				Usage:    "vcs provider. options: 'gitlab'/'github'/'bitbucket'/'repository'",
			},
			&cli.StringFlag{
				Name:    "token",
				Aliases: []string{"t"},
				Usage:   "token for vcs access.",
				EnvVars: []string{"VCS_TOKEN", "GITLAB_TOKEN", "GITHUB_TOKEN"},
			},
			&cli.StringFlag{
				Name:    "object",
				Aliases: []string{"u"},
				Usage:   "repository|org|group to scrape. If not specified all reachable repositories will be collected.",
			},
			&cli.IntFlag{
				Name:  "cloners",
				Value: DefaultClonerN,
				Usage: "number of cloners, more cloners means more memory usage",
			},
			&cli.StringFlag{
				Name:  "provider-url",
				Usage: "base URL of the Git provider API. If not set, defaults URL are used.",
			},
		},
		Action: mainAction,
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func mainAction(c *cli.Context) error {
	if c.Bool("verbose") {
		log.SetLevel(log.InfoLevel)
	} else {
		log.SetLevel(log.ErrorLevel)
	}

	output := os.Stdout

	if c.String("output") != "-" {
		changedOutput, err := os.OpenFile(c.String("output"), os.O_RDWR|os.O_CREATE, os.ModePerm)
		if err != nil {
			return cli.Exit(fmt.Sprintf("Could not open output file: %s", err), 1)
		}

		output = changedOutput

		defer output.Close()
	}

	var srcCloner cloner.Cloner = cloner.NewDiskCloner(c.String("clone-dir"))

	providerOptions := provider.Options{
		OmitForks:    !c.Bool("extract-forks"),
		SkipArchived: c.Bool("skip-archived"),
		BaseURL:      c.String("provider-url"),
	}

	defer func() {
		if r := recover(); r != nil {
			log.Errorln(r)
		}
	}()

	srcProvider, err := getProvider(c.String("provider"), c.String("token"), providerOptions)
	if err != nil {
		cli.ShowAppHelpAndExit(c, 1)
	}

	outputExporter, err := getExporter(c.String("export-format"), output)
	if err != nil {
		cli.ShowAppHelpAndExit(c, 1)
	}

	pipeline := srcfingerprint.Pipeline{
		Provider:     srcProvider,
		Cloner:       srcCloner,
		Analyzer:     &srcfingerprint.Analyzer{},
		ClonersCount: c.Int("cloners"),
	}

	ticker := time.Tick(1 * time.Second)

	eventChannel := runExtract(&pipeline, c.String("object"), c.String("after"))

	// runtime stats
	var (
		totalRepo     int
		doneRepo      int
		gitFilesCount int
	)

	authors := make(map[string]*authorInfo)

loop:
	for {
		select {
		case event, opened := <-eventChannel:
			if !opened {
				break loop
			}

			switch typedEvent := event.(type) {
			case srcfingerprint.RepositoryListPipelineEvent:
				totalRepo = len(typedEvent.Repositories)
			case srcfingerprint.RepositoryPipelineEvent:
				if typedEvent.Finished {
					doneRepo++
				}
			case srcfingerprint.ResultCommitPipelineEvent:
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
			// Collecting gitFiles
			case srcfingerprint.ResultGitFilePipelineEvent:
				gitFilesCount++
				err := outputExporter.AddElement(&exporter.ExportGitFile{
					RepositoryName:    typedEvent.Repository.GetName(),
					RepositoryPrivate: typedEvent.Repository.GetPrivate(),
					GitFile:           *typedEvent.GitFile,
				})

				if err != nil {
					log.Warnln("unable to export git file", err)
				}
			}
		case <-ticker:
			if totalRepo == 0 {
				continue
			}

			log.Infof("%v/%v repos: %v files analyzed\n",
				doneRepo, totalRepo, gitFilesCount)
		}
	}

	log.Infof("Final stats:\n%v/%v repos: %v files analyzed\n",
		doneRepo, totalRepo, gitFilesCount)
	log.Infof("Dumping to output %v\n", c.String("output"))

	if err := outputExporter.Close(); err != nil {
		log.Errorln("Could not save output", err)
	}

	log.Infoln("Done")
	return nil
}
