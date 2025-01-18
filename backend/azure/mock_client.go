package azure

import (
	"io"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/bloberror"

	"github.com/c2fo/vfs/v6"
)

// MockAzureClient is a mock implementation of azure.Client.
type MockAzureClient struct {
	PropertiesError   error
	PropertiesResult  *BlobProperties
	ExpectedError     error
	ExpectedResult    any
	UploadContentType string
}

// Properties returns a PropertiesResult if it exists, otherwise it will return the value of PropertiesError
func (c *MockAzureClient) Properties(_, _ string) (*BlobProperties, error) {
	if c.PropertiesResult == nil {
		return nil, c.PropertiesError
	}
	return c.PropertiesResult, c.PropertiesError
}

// SetMetadata returns the value of ExpectedError
func (c *MockAzureClient) SetMetadata(vfs.File, map[string]*string) error {
	return c.ExpectedError
}

// Upload returns the value of ExpectedError
func (c *MockAzureClient) Upload(_ vfs.File, _ io.ReadSeeker, contentType string) error {
	c.UploadContentType = contentType
	return c.ExpectedError
}

// Download returns ExpectedResult if it exists, otherwise it returns ExpectedError
func (c *MockAzureClient) Download(vfs.File) (io.ReadCloser, error) {
	if c.ExpectedResult != nil {
		return c.ExpectedResult.(io.ReadCloser), nil
	}
	return nil, c.ExpectedError
}

// Copy returns the value of ExpectedError
func (c *MockAzureClient) Copy(_, _ vfs.File) error {
	return c.ExpectedError
}

// List returns the value of ExpectedResult if it exists, otherwise it returns ExpectedError.
func (c *MockAzureClient) List(vfs.Location) ([]string, error) {
	if c.ExpectedResult != nil {
		return c.ExpectedResult.([]string), nil
	}
	return nil, c.ExpectedError
}

// Delete returns the value of ExpectedError
func (c *MockAzureClient) Delete(vfs.File) error {
	return c.ExpectedError
}

// DeleteAllVersions returns the value of ExpectedError
func (c *MockAzureClient) DeleteAllVersions(vfs.File) error {
	return c.ExpectedError
}

var blobNotFoundErr = &azcore.ResponseError{ErrorCode: string(bloberror.BlobNotFound)}
