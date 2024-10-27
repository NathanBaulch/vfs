package vfssimple

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/suite"
)

func TestVFSSimple(t *testing.T) {
	suite.Run(t, &vfsSimpleSuite{})
}

type vfsSimpleSuite struct {
	suite.Suite
}

func (s *vfsSimpleSuite) TestParseURI() {
	tests := []struct {
		uri, message, scheme, authority, path string
		err                                   error
	}{
		{
			uri:     "",
			err:     ErrBlankURI,
			message: "cannot use an empty uri",
		},
		{
			uri:     "asdf@asdf.com",
			err:     ErrMissingScheme,
			message: "email address is not a uri",
		},
		{
			uri:     "1",
			err:     ErrMissingScheme,
			message: "integer is not a uri",
		},
		{
			uri:     "host.com/path",
			err:     ErrMissingScheme,
			message: "missing scheme",
		},
		{
			uri:     "s3.test.com",
			err:     ErrMissingScheme,
			message: "resembles, but is not, a uri",
		},
		{
			uri:     "/some/path/to/file.txt",
			err:     ErrMissingScheme,
			message: "path-only is not a uri",
		},
		{
			uri:     "s3://",
			err:     ErrMissingAuthority,
			message: "scheme only is not a uri without authority",
		},
		{
			uri:     "\u007f",
			err:     errors.New("net/url: invalid control character in URL"),
			message: "invalid char causes parse error",
		},
		{
			uri:       "fake://host.com/path/to/file.txt",
			err:       nil,
			message:   "valid uri for fake scheme",
			scheme:    "fake",
			authority: "host.com",
			path:      "/path/to/file.txt",
		},
		{
			uri:       "file:///path/to/file.txt",
			err:       nil,
			message:   "valid file uri, no authority required",
			scheme:    "file",
			authority: "",
			path:      "/path/to/file.txt",
		},
		{
			uri:       "file://c:/path/to/file.txt",
			err:       nil,
			message:   "valid file uri with authority(volume)",
			scheme:    "file",
			authority: "c:",
			path:      "/path/to/file.txt",
		},
		{
			uri:       "file://c/path/to/file.txt",
			err:       nil,
			message:   "valid file uri with authority(volume), no colon",
			scheme:    "file",
			authority: "c",
			path:      "/path/to/file.txt",
		},
		{
			uri:       "mem:///path/to/file.txt",
			err:       nil,
			message:   "valid mem uri, no authority (namespace) required",
			scheme:    "mem",
			authority: "",
			path:      "/path/to/file.txt",
		},
		{
			uri:       "mem://namespace/path/to/file.txt",
			err:       nil,
			message:   "valid mem uri with namespace(authority)",
			scheme:    "mem",
			authority: "namespace",
			path:      "/path/to/file.txt",
		},
		{
			uri:       "s3://mybucket/path/to/file.txt",
			err:       nil,
			message:   "valid s3 uri",
			scheme:    "s3",
			authority: "mybucket",
			path:      "/path/to/file.txt",
		},
		{
			uri:       "gs://mybucket/path/to/file.txt",
			err:       nil,
			message:   "valid gs uri",
			scheme:    "gs",
			authority: "mybucket",
			path:      "/path/to/file.txt",
		},
		{
			uri:       "https://myaccount.blob.core.windows.net/mycontainer/path/to/file.txt",
			err:       nil,
			message:   "valid azure uri",
			scheme:    "https",
			authority: "mycontainer",
			path:      "/path/to/file.txt",
		},
		{
			uri:       "sftp://user@host.com/path/to/file.txt",
			err:       nil,
			message:   "valid sftp uri",
			scheme:    "sftp",
			authority: "user@host.com",
			path:      "/path/to/file.txt",
		},
		{
			uri:       "sftp://user@host.com:22/path/to/file.txt",
			err:       nil,
			message:   "valid sftp uri, with port",
			scheme:    "sftp",
			authority: "user@host.com:22",
			path:      "/path/to/file.txt",
		},
		{
			uri:       `sftp://domain.com%5Cuser@host.com:22/path/to/file.txt`,
			err:       nil,
			message:   "valid sftp uri, with percent-encoded char",
			scheme:    "sftp",
			authority: `domain.com%5Cuser@host.com:22`,
			path:      "/path/to/file.txt",
		},
		{
			uri:     `sftp://domain.com\user@host.com:22/path/to/file.txt`,
			err:     errors.New("net/url: invalid userinfo"),
			message: `invalid sftp uri, with raw reserved char \`,
		},
	}

	for _, test := range tests {
		s.Run(test.message, func() {
			scheme, authority, path, err := parseURI(test.uri)
			if test.err != nil {
				s.Error(err, test.message)
				if errors.Is(err, test.err) {
					s.True(errors.Is(err, test.err), test.message)
				} else {
					// this is necessary since we can't recreate sentinel errors from url.Parse() to do errors.Is() comparison
					s.Contains(err.Error(), test.err.Error(), test.message)
				}
			} else {
				s.NoError(err, test.message)
				s.Equal(test.scheme, scheme, test.message)
				s.Equal(test.authority, authority, test.message)
				s.Equal(test.path, path, test.message)
			}
		})
	}
}
