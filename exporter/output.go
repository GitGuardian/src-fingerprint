package exporter

import (
	"encoding/json"
	"io"
	"srcfingerprint"
)

type ExportGitFile struct {
	RepositoryName    string `json:"repository_name"` // nolint
	RepositoryPrivate bool   `json:"private"`
	srcfingerprint.GitFile
}

type Exporter interface {
	AddElement(gitFile *ExportGitFile) error
	Close() error
}

type JSONExporter struct {
	elements []*ExportGitFile
	encoder  *json.Encoder
}

func NewJSONExporter(output io.Writer) Exporter {
	return &JSONExporter{
		elements: []*ExportGitFile{},
		encoder:  json.NewEncoder(output),
	}
}

func (e *JSONExporter) AddElement(gitFile *ExportGitFile) error {
	e.elements = append(e.elements, gitFile)

	return nil
}

func (e *JSONExporter) Close() error {
	return e.encoder.Encode(e.elements)
}

type JSONLExporter struct {
	encoder *json.Encoder
}

func NewJSONLExporter(output io.Writer) Exporter {
	return &JSONLExporter{
		encoder: json.NewEncoder(output),
	}
}

func (e *JSONLExporter) AddElement(gitFile *ExportGitFile) error {
	return e.encoder.Encode(gitFile)
}

func (e *JSONLExporter) Close() error {
	return nil
}
