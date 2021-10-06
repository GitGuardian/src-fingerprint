package extractor

type Extractor interface {
	Next() (*GitFile, bool)
	Run(path string, after string)
}

type GitFile struct {
	Sha      string `json:"sha"`
	Type     string `json:"type"`
	Filepath string `json:"filepath"`
	Size     string `json:"size"`
}

type Maker interface {
	Make() Extractor
}
