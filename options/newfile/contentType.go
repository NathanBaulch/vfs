package newfile

import (
	"github.com/c2fo/vfs/v6/options"
	"github.com/c2fo/vfs/v6/utils"
)

const optionNameNewFileContentType = "newFileContentType"

// WithContentType returns ContentType implementation of NewFileOption
func WithContentType(contentType string) options.NewFileOption {
	return utils.Ptr(ContentType(contentType))
}

// ContentType represents the NewFileOption that is used to explicitly specify a content type on created files.
type ContentType string

// NewFileOptionName returns the name of ContentType option
func (*ContentType) NewFileOptionName() string {
	return optionNameNewFileContentType
}
