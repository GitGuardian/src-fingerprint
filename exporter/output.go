package exporter

import (
	"compress/gzip"
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
	writer   io.WriteCloser
}

func NewJSONExporter(output io.WriteCloser) Exporter {
	return &JSONExporter{
		elements: []*ExportGitFile{},
		encoder:  json.NewEncoder(output),
		writer:   output,
	}
}

func NewGzipJSONExporter(output io.Writer) Exporter {
	compressedWriter := gzip.NewWriter(output)

	return &JSONExporter{
		elements: []*ExportGitFile{},
		encoder:  json.NewEncoder(compressedWriter),
		writer:   compressedWriter,
	}
}

func (e *JSONExporter) AddElement(gitFile *ExportGitFile) error {
	e.elements = append(e.elements, gitFile)

	return nil
}

func (e *JSONExporter) Close() error {
	if err1 := e.encoder.Encode(e.elements); err1 != nil {
		return err1
	}

	if err2 := e.writer.Close(); err2 != nil {
		return err2
	}

	return nil
}

type JSONLExporter struct {
	encoder *json.Encoder
	writer  io.WriteCloser
}

func NewJSONLExporter(output io.WriteCloser) Exporter {
	return &JSONLExporter{
		encoder: json.NewEncoder(output),
		writer:  output,
	}
}

func NewGzipJSONLExporter(output io.Writer) Exporter {
	compressedWriter := gzip.NewWriter(output)

	return &JSONLExporter{
		encoder: json.NewEncoder(compressedWriter),
		writer:  compressedWriter,
	}
}

func (e *JSONLExporter) AddElement(gitFile *ExportGitFile) error {
	return e.encoder.Encode(gitFile)
}

func (e *JSONLExporter) Close() error {
	return e.writer.Close()
}
