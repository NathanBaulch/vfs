package ftp

import (
	"context"
	"errors"
	"path"

	"github.com/c2fo/vfs/v6"
	"github.com/c2fo/vfs/v6/backend"
	"github.com/c2fo/vfs/v6/backend/ftp/types"
	"github.com/c2fo/vfs/v6/options"
	"github.com/c2fo/vfs/v6/utils"
)

// Scheme defines the filesystem type.
const (
	Scheme = "ftp"
	name   = "File Transfer Protocol"
)

// FileSystem implements vfs.FileSystem for the FTP filesystem.
type FileSystem struct {
	options         *Options
	ftpclient       types.Client
	dataconn        types.DataConn
	clientFactory   clientFactory
	dataConnFactory dataConnFactory
}

// Retry will return the default no-op retrier. The FTP client provides its own retryer interface, and is available
// to override via the ftp.FileSystem Options type.
func (fs *FileSystem) Retry() vfs.Retry {
	return vfs.DefaultRetryer()
}

// NewFile function returns the FTP implementation of vfs.File.
func (fs *FileSystem) NewFile(authority, filePath string, opts ...options.NewFileOption) (vfs.File, error) {
	if fs == nil {
		return nil, errors.New("non-nil ftp.FileSystem pointer is required")
	}
	if filePath == "" {
		return nil, errors.New("non-empty string for path is required")
	}
	if err := utils.ValidateAbsoluteFilePath(filePath); err != nil {
		return nil, err
	}

	auth, err := utils.NewAuthority(authority)
	if err != nil {
		return nil, err
	}

	return &File{
		fileSystem: fs,
		authority:  auth,
		path:       path.Clean(filePath),
		opts:       opts,
	}, nil
}

// NewLocation function returns the FTP implementation of vfs.Location.
func (fs *FileSystem) NewLocation(authority, locPath string) (vfs.Location, error) {
	if fs == nil {
		return nil, errors.New("non-nil ftp.FileSystem pointer is required")
	}
	if err := utils.ValidateAbsoluteLocationPath(locPath); err != nil {
		return nil, err
	}

	auth, err := utils.NewAuthority(authority)
	if err != nil {
		return nil, err
	}

	return &Location{
		fileSystem: fs,
		path:       utils.EnsureTrailingSlash(path.Clean(locPath)),
		Authority:  auth,
	}, nil
}

// Name returns "Secure File Transfer Protocol"
func (fs *FileSystem) Name() string {
	return name
}

// Scheme return "ftp" as the initial part of a file URI ie: ftp://
func (fs *FileSystem) Scheme() string {
	return Scheme
}

// DataConn returns the underlying ftp data connection, creating it, if necessary
// See Overview for authentication resolution
func (fs *FileSystem) DataConn(ctx context.Context, authority utils.Authority, t types.OpenType, f *File) (types.DataConn, error) {
	if t != types.SingleOp && f == nil {
		return nil, errors.New("can not create DataConn for read or write for a nil file")
	}

	if fs == nil {
		return nil, errors.New("can not get a dataconn for a nil fileset")
	}
	if fs.dataconn != nil && fs.dataconn.Mode() != t {
		// wrong session type ... close current session and unset it (ps so we can set a new one after)
		if err := fs.dataconn.Close(); err != nil {
			return nil, err
		}
		fs.dataconn = nil
	}

	if fs.dataconn == nil {
		client, err := fs.Client(ctx, authority)
		if err != nil {
			return nil, err
		}

		fs.dataconn, err = fs.dataConnFactory(client, f, t)
		if err != nil {
			return nil, err
		}
	}

	return fs.dataconn, nil
}

// Client returns the underlying ftp client, creating it, if necessary
// See Overview for authentication resolution
func (fs *FileSystem) Client(ctx context.Context, authority utils.Authority) (types.Client, error) {
	if fs.ftpclient == nil {
		var err error
		fs.ftpclient, err = fs.clientFactory(ctx, authority, fs.options)
		if err != nil {
			return nil, err
		}
	}
	return fs.ftpclient, nil
}

// WithOptions sets options for client and returns the filesystem (chainable)
func (fs *FileSystem) WithOptions(opts vfs.Options) *FileSystem {
	// only set options if vfs.Options is ftp.Options
	switch opts := opts.(type) {
	case Options:
		fs.options = &opts
	case *Options:
		fs.options = opts
	default:
		return fs
	}
	// we set client to nil to ensure that a new client is created
	fs.ftpclient = nil
	return fs
}

// WithClient passes in an ftp client and returns the filesystem (chainable)
func (fs *FileSystem) WithClient(client types.Client) *FileSystem {
	fs.ftpclient = client
	fs.options = nil
	return fs
}

type clientFactory func(ctx context.Context, authority utils.Authority, opts *Options) (types.Client, error)

type dataConnFactory func(client types.Client, f *File, t types.OpenType) (types.DataConn, error)

// NewFileSystem initializer for fileSystem struct.
func NewFileSystem() *FileSystem {
	return &FileSystem{
		clientFactory:   getClient,
		dataConnFactory: getDataConn,
	}
}

func init() {
	// registers a default FileSystem
	backend.Register(Scheme, NewFileSystem())
}
