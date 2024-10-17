package newfile

import "github.com/c2fo/vfs/v6/options"

const optionNameNewFileContentType = "newFileContentType"

func WithContentType(contentType string) options.NewFileOption {
	ct := ContentType(contentType)
	return &ct
}

type ContentType string

func (ct *ContentType) NewFileOptionName() string {
	return optionNameNewFileContentType
}
