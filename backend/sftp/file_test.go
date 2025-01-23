package sftp

import (
	"bytes"
	"errors"
	"io"
	"os"
	"path"
	"strings"
	"testing"
	"time"

	"github.com/pkg/sftp"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/c2fo/vfs/v6"
	"github.com/c2fo/vfs/v6/backend/sftp/mocks"
	vfsmocks "github.com/c2fo/vfs/v6/mocks"
	"github.com/c2fo/vfs/v6/utils"
)

type fileTestSuite struct {
	suite.Suite
	sftpMock *mocks.Client
	fs       FileSystem
	testFile vfs.File
}

func (ts *fileTestSuite) SetupTest() {
	var err error
	ts.sftpMock = mocks.NewClient(ts.T())
	ts.fs = FileSystem{sftpclient: ts.sftpMock}
	ts.testFile, err = ts.fs.NewFile("user@host.com:22", "/some/path/to/file.txt")
	ts.Require().NoError(err, "Shouldn't return error creating test sftp.File instance.")
}

// this wraps strings.Reader to satisfy readWriteSeekCloser interface
type nopWriteCloser struct {
	io.ReadSeeker
}

func (nopWriteCloser) Close() error                      { return nil }
func (nopWriteCloser) Write(_ []byte) (n int, err error) { return 0, nil }

func (ts *fileTestSuite) TestRead() {
	// set up sftpfile
	filepath := "/some/path.txt"
	client := mocks.NewClient(ts.T())

	contents := "hello world!"
	auth, err := utils.NewAuthority("user@host1.com:22")
	ts.Require().NoError(err)
	sftpfile := &File{
		fileSystem: &FileSystem{
			sftpclient: client,
		},
		Authority: auth,
		path:      filepath,
		sftpfile:  &nopWriteCloser{strings.NewReader(contents)},
	}
	// perform test
	localFile := &bytes.Buffer{}

	buffer := make([]byte, utils.TouchCopyMinBufferSize)
	b, err := io.CopyBuffer(localFile, sftpfile, buffer)
	ts.Require().NoError(err, "no error expected")
	ts.Equal(int64(12), b, "byte count after copy")
	ts.Require().NoError(sftpfile.Close(), "no error expected")
	ts.Equal(localFile.String(), contents, "Copying an sftp file to a buffer should fill buffer with localfile's contents")
}

func (ts *fileTestSuite) TestSeek() {
	// set up sftpfile
	filepath := "/some/path.txt"
	client := mocks.NewClient(ts.T())

	contents := "hello world!"
	auth, err := utils.NewAuthority("user@host1.com:22")
	ts.Require().NoError(err)

	sftpfile := &File{
		fileSystem: &FileSystem{
			sftpclient: client,
		},
		Authority: auth,
		path:      filepath,
		sftpfile:  &nopWriteCloser{strings.NewReader(contents)},
	}
	// perform test
	_, err = sftpfile.Seek(6, io.SeekStart)
	ts.Require().NoError(err, "no error expected")

	localFile := &bytes.Buffer{}

	buffer := make([]byte, utils.TouchCopyMinBufferSize)
	_, err = io.CopyBuffer(localFile, sftpfile, buffer)
	ts.Require().NoError(err, "no error expected")

	ts.Equal("world!", localFile.String(), "Seeking should move the sftp file cursor as expected")

	localFile = &bytes.Buffer{}
	_, err = sftpfile.Seek(0, io.SeekStart)
	ts.Require().NoError(err, "no error expected")

	buffer = make([]byte, utils.TouchCopyMinBufferSize)
	_, copyErr2 := io.CopyBuffer(localFile, sftpfile, buffer)
	ts.Require().NoError(copyErr2, "no error expected")
	ts.Equal(contents, localFile.String(), "Subsequent calls to seek work on temp sftp file as expected")

	err = sftpfile.Close()
	ts.Require().NoError(err, "no error expected")
}

func (ts *fileTestSuite) Test_openFile() {
	testCases := []struct {
		name           string
		flags          int
		setupMocks     func(client *mocks.Client)
		expectedError  bool
		expectedErrMsg string
	}{
		{
			name:  "Open file for read",
			flags: os.O_RDONLY,
			setupMocks: func(client *mocks.Client) {
				client.EXPECT().OpenFile("/some/path.txt", os.O_RDONLY).Return(&sftp.File{}, nil)
			},
			expectedError: false,
		},
		{
			name:  "Open file for write",
			flags: os.O_WRONLY | os.O_CREATE,
			setupMocks: func(client *mocks.Client) {
				client.EXPECT().MkdirAll("/some").Return(nil)
				client.EXPECT().Chmod("/some/path.txt", os.FileMode(0o644)).Return(nil)
				client.EXPECT().OpenFile("/some/path.txt", os.O_WRONLY|os.O_CREATE).Return(&sftp.File{}, nil)
			},
			expectedError: false,
		},
		{
			name:  "Open file for create",
			flags: os.O_RDWR | os.O_CREATE,
			setupMocks: func(client *mocks.Client) {
				client.EXPECT().MkdirAll(path.Dir("/some/path.txt")).Return(nil)
				client.EXPECT().OpenFile("/some/path.txt", os.O_RDWR|os.O_CREATE).Return(&sftp.File{}, nil)
			},
			expectedError: false,
		},
		{
			name:  "Open file for create with error",
			flags: os.O_RDWR | os.O_CREATE,
			setupMocks: func(client *mocks.Client) {
				client.EXPECT().MkdirAll(path.Dir("/some/path.txt")).Return(errors.New("mkdir error"))
			},
			expectedError:  true,
			expectedErrMsg: "mkdir error",
		},
		{
			name:  "Open file with default permissions",
			flags: os.O_WRONLY,
			setupMocks: func(client *mocks.Client) {
				client.EXPECT().OpenFile("/some/path.txt", os.O_WRONLY).Return(&sftp.File{}, nil)
				client.EXPECT().Chmod("/some/path.txt", os.FileMode(0o644)).Return(nil)
			},
			expectedError: false,
		},
		{
			name:  "Open file with default permissions error",
			flags: os.O_WRONLY,
			setupMocks: func(client *mocks.Client) {
				client.EXPECT().OpenFile("/some/path.txt", os.O_WRONLY).Return(&sftp.File{}, nil)
				client.EXPECT().Chmod("/some/path.txt", os.FileMode(0o644)).Return(errors.New("chmod error"))
			},
			expectedError:  true,
			expectedErrMsg: "chmod error",
		},
	}

	for _, tc := range testCases {
		ts.Run(tc.name, func() {
			client := mocks.NewClient(ts.T())
			tc.setupMocks(client)

			authority, err := utils.NewAuthority("sftp://user@host:22")
			ts.Require().NoError(err)
			file := &File{
				path:      "/some/path.txt",
				Authority: authority,
				fileSystem: &FileSystem{
					sftpclient: client,
					options:    &Options{FilePermissions: utils.Ptr("0644")},
				},
			}

			_, err = file._open(tc.flags)
			if tc.expectedError {
				ts.Require().Error(err)
				ts.Contains(err.Error(), tc.expectedErrMsg)
			} else {
				ts.Require().NoError(err)
			}
		})
	}
}

func (ts *fileTestSuite) TestExists() {
	sftpfile, err := ts.fs.NewFile("user@host.com", "/path/hello.txt")
	ts.Require().NoError(err, "Shouldn't fail creating new file.")

	ts.sftpMock.EXPECT().Stat(sftpfile.Path()).Return(nil, nil).Once()

	exists, err := sftpfile.Exists()
	ts.Require().NoError(err, "Shouldn't return an error when exists is true")
	ts.True(exists, "Should return true for exists based on this setup")
}

func (ts *fileTestSuite) TestNotExists() {
	sftpfile, err := ts.fs.NewFile("user@host.com", "/path/hello.txt")
	ts.Require().NoError(err, "Shouldn't fail creating new file.")

	ts.sftpMock.EXPECT().Stat(sftpfile.Path()).Return(nil, os.ErrNotExist).Once()
	exists, err := sftpfile.Exists()
	ts.Require().NoError(err, "Error from key not existing should be hidden since it just confirms it doesn't")
	ts.False(exists, "Should return false for exists based on setup")
}

func (ts *fileTestSuite) TestCopyToFile() {
	content := "this is a test"

	// set up source
	sourceClient := mocks.NewClient(ts.T())

	sourceSftpFile := mocks.NewReadWriteSeekCloser(ts.T())

	sourceSftpFile.EXPECT().Read(mock.Anything).Return(len(content), nil).Once()
	sourceSftpFile.EXPECT().Read(mock.Anything).Return(0, io.EOF).Once()
	sourceSftpFile.EXPECT().Close().Return(nil).Once()

	auth, err := utils.NewAuthority("user@host1.com:22")
	ts.Require().NoError(err)

	sourceFile := &File{
		fileSystem: &FileSystem{
			sftpclient: sourceClient,
		},
		Authority: auth,
		path:      "/some/path.txt",
		sftpfile:  sourceSftpFile,
	}

	// set up target
	targetClient := mocks.NewClient(ts.T())

	targetSftpFile := mocks.NewReadWriteSeekCloser(ts.T())
	targetSftpFile.EXPECT().Write(mock.Anything).Return(len(content), nil).Once()
	targetSftpFile.EXPECT().Close().Return(nil).Once()

	auth2, err := utils.NewAuthority("user@host2.com:22")
	ts.Require().NoError(err)

	targetFile := &File{
		fileSystem: &FileSystem{
			sftpclient: targetClient,
		},
		Authority: auth2,
		path:      "/some/path.txt",
		sftpfile:  targetSftpFile,
		opener:    func(Client, string, int) (readWriteSeekCloser, error) { return targetSftpFile, nil },
	}

	// run tests
	err = sourceFile.CopyToFile(targetFile)
	ts.Require().NoError(err, "Error shouldn't be returned from successful call to CopyToFile")
}

func (ts *fileTestSuite) TestCopyToFileBuffered() {
	content := "this is a test"

	// set up source
	sourceClient := mocks.NewClient(ts.T())

	sourceSftpFile := mocks.NewReadWriteSeekCloser(ts.T())

	sourceSftpFile.EXPECT().Read(mock.Anything).Return(len(content), nil).Once()
	sourceSftpFile.EXPECT().Read(mock.Anything).Return(0, io.EOF).Once()
	sourceSftpFile.EXPECT().Close().Return(nil).Once()

	auth, err := utils.NewAuthority("user@host1.com:22")
	ts.Require().NoError(err)

	sourceFile := &File{
		fileSystem: &FileSystem{
			sftpclient: sourceClient,
			options:    &Options{FileBufferSize: 2 * utils.TouchCopyMinBufferSize},
		},
		Authority: auth,
		path:      "/some/path.txt",
		sftpfile:  sourceSftpFile,
	}

	// set up target
	targetClient := mocks.NewClient(ts.T())

	targetSftpFile := mocks.NewReadWriteSeekCloser(ts.T())
	targetSftpFile.EXPECT().Write(mock.Anything).Return(len(content), nil).Once()
	targetSftpFile.EXPECT().Close().Return(nil).Once()

	auth2, err := utils.NewAuthority("user@host2.com:22")
	ts.Require().NoError(err)

	targetFile := &File{
		fileSystem: &FileSystem{
			sftpclient: targetClient,
		},
		Authority: auth2,
		path:      "/some/path.txt",
		sftpfile:  targetSftpFile,
		opener:    func(Client, string, int) (readWriteSeekCloser, error) { return targetSftpFile, nil },
	}

	targetMockLocation := &vfsmocks.Location{}
	targetMockLocation.EXPECT().NewFile(mock.Anything).Return(targetFile, nil)

	// run tests
	err = sourceFile.CopyToFile(targetFile)
	ts.Require().NoError(err, "Error shouldn't be returned from successful call to CopyToFile")
}

func (ts *fileTestSuite) TestCopyToFileEmpty() {
	content := ""

	// set up source
	sourceClient := mocks.NewClient(ts.T())

	sourceSftpFile := mocks.NewReadWriteSeekCloser(ts.T())
	sourceSftpFile.EXPECT().Read(mock.Anything).Return(0, io.EOF).Once()
	sourceSftpFile.EXPECT().Close().Return(nil).Once()

	auth, err := utils.NewAuthority("user@host1.com:22")
	ts.Require().NoError(err)

	sourceFile := &File{
		fileSystem: &FileSystem{
			sftpclient: sourceClient,
		},
		Authority: auth,
		path:      "/some/path.txt",
		sftpfile:  sourceSftpFile,
	}

	// set up target
	targetClient := mocks.NewClient(ts.T())

	targetSftpFile := mocks.NewReadWriteSeekCloser(ts.T())
	targetSftpFile.EXPECT().Write(mock.Anything).Return(len(content), nil).Once()
	targetSftpFile.EXPECT().Close().Return(nil).Once()

	auth2, err := utils.NewAuthority("user@host2.com:22")
	ts.Require().NoError(err)

	targetFile := &File{
		fileSystem: &FileSystem{
			sftpclient: targetClient,
		},
		Authority: auth2,
		path:      "/some/path.txt",
		sftpfile:  targetSftpFile,
		opener:    func(Client, string, int) (readWriteSeekCloser, error) { return targetSftpFile, nil },
	}

	targetMockLocation := &vfsmocks.Location{}
	targetMockLocation.EXPECT().NewFile(mock.Anything).Return(targetFile, nil)

	// run tests
	err = sourceFile.CopyToFile(targetFile)
	ts.Require().NoError(err, "Error shouldn't be returned from successful call to CopyToFile")
}

func (ts *fileTestSuite) TestCopyToFileEmptyBuffered() {
	content := ""

	// set up source
	sourceClient := mocks.NewClient(ts.T())

	sourceSftpFile := mocks.NewReadWriteSeekCloser(ts.T())
	sourceSftpFile.EXPECT().Read(mock.Anything).Return(0, io.EOF).Once()
	sourceSftpFile.EXPECT().Close().Return(nil).Once()

	auth, err := utils.NewAuthority("user@host1.com:22")
	ts.Require().NoError(err)

	sourceFile := &File{
		fileSystem: &FileSystem{
			sftpclient: sourceClient,
			options:    &Options{FileBufferSize: 2 * utils.TouchCopyMinBufferSize},
		},
		Authority: auth,
		path:      "/some/path.txt",
		sftpfile:  sourceSftpFile,
	}

	// set up target
	targetClient := mocks.NewClient(ts.T())

	targetSftpFile := mocks.NewReadWriteSeekCloser(ts.T())
	targetSftpFile.EXPECT().Write(mock.Anything).Return(len(content), nil).Once()
	targetSftpFile.EXPECT().Close().Return(nil).Once()

	auth2, err := utils.NewAuthority("user@host2.com:22")
	ts.Require().NoError(err)

	targetFile := &File{
		fileSystem: &FileSystem{
			sftpclient: targetClient,
		},
		Authority: auth2,
		path:      "/some/path.txt",
		sftpfile:  targetSftpFile,
		opener:    func(Client, string, int) (readWriteSeekCloser, error) { return targetSftpFile, nil },
	}

	targetMockLocation := &vfsmocks.Location{}
	targetMockLocation.EXPECT().NewFile(mock.Anything).Return(targetFile, nil)

	// run tests
	err = sourceFile.CopyToFile(targetFile)
	ts.Require().NoError(err, "Error shouldn't be returned from successful call to CopyToFile")
}

func (ts *fileTestSuite) TestCopyToLocation() {
	content := "this is a location test"

	// set up source
	sourceClient := mocks.NewClient(ts.T())

	sourceSftpFile := mocks.NewReadWriteSeekCloser(ts.T())
	sourceSftpFile.EXPECT().Read(mock.Anything).Return(len(content), nil).Once()
	sourceSftpFile.EXPECT().Read(mock.Anything).Return(0, io.EOF).Once()
	sourceSftpFile.EXPECT().Close().Return(nil).Once()

	auth, err := utils.NewAuthority("user@host1.com:22")
	ts.Require().NoError(err)

	sourceFile := &File{
		fileSystem: &FileSystem{
			sftpclient: sourceClient,
		},
		Authority: auth,
		path:      "/some/path.txt",
		sftpfile:  sourceSftpFile,
	}

	// set up target
	targetClient := mocks.NewClient(ts.T())

	targetSftpFile := mocks.NewReadWriteSeekCloser(ts.T())
	targetSftpFile.EXPECT().Write(mock.Anything).Return(len(content), nil).Once()
	targetSftpFile.EXPECT().Close().Return(nil).Once()

	auth2, err := utils.NewAuthority("user@host2.com:22")
	ts.Require().NoError(err)

	targetFile := &File{
		fileSystem: &FileSystem{
			sftpclient: targetClient,
		},
		Authority: auth2,
		path:      "/some/path.txt",
		sftpfile:  targetSftpFile,
		opener:    func(Client, string, int) (readWriteSeekCloser, error) { return targetSftpFile, nil },
	}

	targetMockLocation := &vfsmocks.Location{}
	targetMockLocation.EXPECT().NewFile(mock.Anything).Return(targetFile, nil)

	// run tests
	newFile, err := sourceFile.CopyToLocation(targetMockLocation)
	ts.Require().NoError(err, "Error shouldn't be returned from successful call to CopyToFile")

	ts.Equal("sftp://user@host2.com:22/some/path.txt", newFile.URI(), "new file uri check")
}

func (ts *fileTestSuite) TestMoveToFile_differentAuthority() {
	content := "blah"

	// set up source
	sourceClient := mocks.NewClient(ts.T())
	sourceClient.EXPECT().Remove(mock.Anything).Return(nil).Once()

	sourceSftpFile := mocks.NewReadWriteSeekCloser(ts.T())
	sourceSftpFile.EXPECT().Read(mock.Anything).Return(len(content), nil).Once()
	sourceSftpFile.EXPECT().Read(mock.Anything).Return(0, io.EOF).Once()
	sourceSftpFile.EXPECT().Close().Return(nil).Once()

	auth, err := utils.NewAuthority("user@host1.com:22")
	ts.Require().NoError(err)

	sourceFile := &File{
		fileSystem: &FileSystem{
			sftpclient: sourceClient,
		},
		Authority: auth,
		path:      "/some/path.txt",
		sftpfile:  sourceSftpFile,
	}

	// set up target
	targetClient := mocks.NewClient(ts.T())

	targetSftpFile := mocks.NewReadWriteSeekCloser(ts.T())
	targetSftpFile.EXPECT().Write(mock.Anything).Return(len(content), nil).Once()
	targetSftpFile.EXPECT().Close().Return(nil).Once()

	auth2, err := utils.NewAuthority("user@host2.com:22")
	ts.Require().NoError(err)

	targetFile := &File{
		fileSystem: &FileSystem{
			sftpclient: targetClient,
		},
		Authority: auth2,
		path:      "/some/path.txt",
		sftpfile:  targetSftpFile,
		opener:    func(Client, string, int) (readWriteSeekCloser, error) { return targetSftpFile, nil },
	}

	// run tests
	err = sourceFile.MoveToFile(targetFile)
	ts.Require().NoError(err, "Error shouldn't be returned from successful call to CopyToFile")
}

func (ts *fileTestSuite) TestMoveToFile_sameAuthority() {
	// set up source
	sourceClient := mocks.NewClient(ts.T())
	sourceClient.EXPECT().Rename(mock.Anything, mock.Anything).Return(nil).Once()
	sourceClient.EXPECT().MkdirAll(mock.Anything).Return(nil).Once()

	auth, err := utils.NewAuthority("user@host1.com:22")
	ts.Require().NoError(err)

	sourceFile := &File{
		fileSystem: &FileSystem{
			sftpclient: sourceClient,
		},
		Authority: auth,
		path:      "/some/path.txt",
	}

	rws := mocks.NewReadWriteSeekCloser(ts.T())
	sourceFile.opener = func(Client, string, int) (readWriteSeekCloser, error) { return rws, nil }

	// set up target
	targetClient := mocks.NewClient(ts.T())
	targetClient.EXPECT().Stat(mock.Anything).Return(nil, os.ErrNotExist).Twice()

	auth2, err := utils.NewAuthority("user@host1.com:22")
	ts.Require().NoError(err)

	targetFile := &File{
		fileSystem: &FileSystem{
			sftpclient: targetClient,
		},
		Authority: auth2,
		path:      "/some/other/path.txt",
	}

	// run tests
	err = sourceFile.MoveToFile(targetFile)
	ts.Require().NoError(err, "Error shouldn't be returned from successful call to CopyToFile")
}

func (ts *fileTestSuite) TestMoveToFile_fileExists() {
	// set up source
	sourceClient := mocks.NewClient(ts.T())

	sourceClient.EXPECT().Rename(mock.Anything, mock.Anything).Return(nil).Once()
	sourceClient.EXPECT().MkdirAll(mock.Anything).Return(nil).Once()

	auth, err := utils.NewAuthority("user@host1.com:22")
	ts.Require().NoError(err)

	sourceFile := &File{
		fileSystem: &FileSystem{
			sftpclient: sourceClient,
		},
		Authority: auth,
		path:      "/some/path.txt",
	}

	rws := mocks.NewReadWriteSeekCloser(ts.T())
	sourceFile.opener = func(Client, string, int) (readWriteSeekCloser, error) { return rws, nil }

	// set up target
	targetFileInfo := mocks.NewFileInfo(ts.T())

	targetClient := mocks.NewClient(ts.T())

	auth2, err := utils.NewAuthority("user@host1.com:22")
	ts.Require().NoError(err)

	targetFile := &File{
		fileSystem: &FileSystem{
			sftpclient: targetClient,
		},
		Authority: auth2,
		path:      "/some/other/path.txt",
	}
	targetClient.EXPECT().Stat(targetFile.Location().Path()).Return(nil, os.ErrNotExist).Once()
	targetClient.EXPECT().Stat(targetFile.path).Return(targetFileInfo, nil).Once()
	targetClient.EXPECT().Remove(targetFile.path).Return(nil).Once()

	// run tests
	err = sourceFile.MoveToFile(targetFile)
	ts.Require().NoError(err, "Error shouldn't be returned from successful call to CopyToFile")
}

func (ts *fileTestSuite) TestMoveToLocation() {
	content := "loc test"

	// set up source
	sourceClient := mocks.NewClient(ts.T())
	sourceClient.EXPECT().Remove(mock.Anything).Return(nil).Once()

	sourceSftpFile := mocks.NewReadWriteSeekCloser(ts.T())
	sourceSftpFile.EXPECT().Read(mock.Anything).Return(len(content), nil).Once()
	sourceSftpFile.EXPECT().Read(mock.Anything).Return(0, io.EOF).Once()
	sourceSftpFile.EXPECT().Close().Return(nil).Once()

	auth, err := utils.NewAuthority("user@host1.com:22")
	ts.Require().NoError(err)

	sourceFile := &File{
		fileSystem: &FileSystem{
			sftpclient: sourceClient,
		},
		Authority: auth,
		path:      "/some/path.txt",
		sftpfile:  sourceSftpFile,
	}

	// set up target
	targetClient := mocks.NewClient(ts.T())

	targetSftpFile := mocks.NewReadWriteSeekCloser(ts.T())
	targetSftpFile.EXPECT().Write(mock.Anything).Return(len(content), nil).Once()
	targetSftpFile.EXPECT().Close().Return(nil).Once()

	auth2, err := utils.NewAuthority("user@host2.com:22")
	ts.Require().NoError(err)

	targetFile := &File{
		fileSystem: &FileSystem{
			sftpclient: targetClient,
		},
		Authority: auth2,
		path:      "/some/other/path.txt",
		sftpfile:  targetSftpFile,
		opener:    func(Client, string, int) (readWriteSeekCloser, error) { return targetSftpFile, nil },
	}

	targetMockLocation := &vfsmocks.Location{}
	targetMockLocation.EXPECT().NewFile(mock.Anything).Return(targetFile, nil)

	// run tests
	newFile, err := sourceFile.MoveToLocation(targetMockLocation)
	ts.Require().NoError(err, "Error shouldn't be returned from successful call to CopyToFile")

	ts.Equal("sftp://user@host2.com:22/some/other/path.txt", newFile.URI(), "new file uri check")
}

func (ts *fileTestSuite) TestTouch() {
	err := errors.New("some error")
	testCases := []struct {
		name           string
		filePath       string
		fileExists     bool
		setPermissions bool
		expectedError  error
		setupMocks     func(client *mocks.Client, sftpFile *mocks.ReadWriteSeekCloser, fileInfo *mocks.FileInfo)
	}{
		{
			name:       "file exists",
			filePath:   "/some/path.txt",
			fileExists: true,
			setupMocks: func(client *mocks.Client, _ *mocks.ReadWriteSeekCloser, fileInfo *mocks.FileInfo) {
				client.EXPECT().Stat("/some/path.txt").Return(fileInfo, nil).Once()
				client.EXPECT().Chtimes("/some/path.txt", mock.Anything, mock.Anything).Return(nil).Once()
			},
		},
		{
			name:       "file does not exist",
			filePath:   "/some/path.txt",
			fileExists: false,
			setupMocks: func(client *mocks.Client, sftpFile *mocks.ReadWriteSeekCloser, _ *mocks.FileInfo) {
				client.EXPECT().Stat("/some/path.txt").Return(nil, os.ErrNotExist).Once()
				sftpFile.EXPECT().Close().Return(nil).Once()
			},
		},
		{
			name:           "set default permissions",
			filePath:       "/some/path.txt",
			fileExists:     true,
			setPermissions: true,
			setupMocks: func(client *mocks.Client, _ *mocks.ReadWriteSeekCloser, fileInfo *mocks.FileInfo) {
				client.EXPECT().Stat("/some/path.txt").Return(fileInfo, nil).Once()
				client.EXPECT().Chmod("/some/path.txt", os.FileMode(0o666)).Return(nil).Once()
				client.EXPECT().Chtimes("/some/path.txt", mock.Anything, mock.Anything).Return(nil).Once()
			},
		},
		{
			name:          "error on stat",
			filePath:      "/some/path.txt",
			expectedError: err,
			setupMocks: func(client *mocks.Client, _ *mocks.ReadWriteSeekCloser, _ *mocks.FileInfo) {
				client.EXPECT().Stat("/some/path.txt").Return(nil, err).Once()
			},
		},
		{
			name:          "error on chtimes",
			filePath:      "/some/path.txt",
			fileExists:    true,
			expectedError: err,
			setupMocks: func(client *mocks.Client, _ *mocks.ReadWriteSeekCloser, fileInfo *mocks.FileInfo) {
				client.EXPECT().Stat("/some/path.txt").Return(fileInfo, nil).Once()
				client.EXPECT().Chtimes("/some/path.txt", mock.Anything, mock.Anything).Return(err).Once()
			},
		},
		{
			name:     "setPermissions returns error",
			filePath: "/some/path.txt",
			setupMocks: func(client *mocks.Client, _ *mocks.ReadWriteSeekCloser, fileInfo *mocks.FileInfo) {
				client.EXPECT().Stat("/some/path.txt").Return(fileInfo, nil).Once()
				client.EXPECT().Chmod("/some/path.txt", os.FileMode(0o666)).Return(err).Once()
			},
			expectedError:  err,
			setPermissions: true,
		},
	}

	for _, tc := range testCases {
		ts.Run(tc.name, func() {
			client := mocks.NewClient(ts.T())
			sftpFile := mocks.NewReadWriteSeekCloser(ts.T())
			fileInfo := mocks.NewFileInfo(ts.T())

			auth, err := utils.NewAuthority("user@host1.com:22")
			ts.Require().NoError(err)

			file := &File{
				fileSystem: &FileSystem{
					sftpclient: client,
					options: &Options{
						FilePermissions: func() *string {
							if tc.setPermissions {
								return utils.Ptr("0666")
							}
							return nil
						}(),
					},
				},
				Authority: auth,
				path:      tc.filePath,
				sftpfile:  sftpFile,
			}

			tc.setupMocks(client, sftpFile, fileInfo)

			err = file.Touch()
			if tc.expectedError != nil {
				ts.Require().ErrorIs(err, tc.expectedError)
			} else {
				ts.Require().NoError(err)
			}
		})
	}
}

func (ts *fileTestSuite) TestDelete() {
	ts.sftpMock.EXPECT().Remove(ts.testFile.Path()).Return(nil).Once()
	err := ts.testFile.Delete()
	ts.Require().NoError(err, "Successful delete should not return an error.")
}

func (ts *fileTestSuite) TestLastModified() {
	now := time.Now()
	file1 := mocks.NewFileInfo(ts.T())
	file1.EXPECT().ModTime().Return(now)
	ts.sftpMock.EXPECT().Stat(ts.testFile.Path()).Return(file1, nil)
	modTime, err := ts.testFile.LastModified()
	ts.Require().NoError(err, "Error should be nil when correctly returning time of object.")
	ts.Equal(&now, modTime, "Returned time matches expected LastModified time.")
}

func (ts *fileTestSuite) TestLastModifiedFail() {
	myErr := errors.New("some error")
	ts.sftpMock.EXPECT().Stat(ts.testFile.Path()).Return(nil, myErr)
	m, err := ts.testFile.LastModified()
	ts.Require().Error(err, "got error as expected")
	ts.Nil(m, "nil ModTime returned")
}

func (ts *fileTestSuite) TestName() {
	ts.Equal("file.txt", ts.testFile.Name(), "Name should return just the name of the file.")
}

func (ts *fileTestSuite) TestSize() {
	contentLength := int64(100)
	file1 := mocks.NewFileInfo(ts.T())
	file1.EXPECT().Size().Return(contentLength)
	ts.sftpMock.EXPECT().Stat(ts.testFile.Path()).Return(file1, nil).Once()
	size, err := ts.testFile.Size()
	ts.Require().NoError(err, "Error should be nil when requesting size for file that exists.")
	ts.Equal(uint64(contentLength), size, "Size should return the ContentLength value from s3 HEAD request.")

	ts.sftpMock.EXPECT().Stat(ts.testFile.Path()).Return(&mocks.FileInfo{}, errors.New("some error")).Once()
	size, err = ts.testFile.Size()
	ts.Require().Error(err, "expect error")
	ts.Zero(size, "Size should be 0 on error")
}

func (ts *fileTestSuite) TestPath() {
	ts.Equal("/some/path/to/file.txt", ts.testFile.Path(), "Should return file.key (with leading slash)")
}

func (ts *fileTestSuite) TestURI() {
	expected := "sftp://user@host.com:22/some/path/to/file.txt"
	ts.Equal(expected, ts.testFile.URI(), "URI test")

	expected = "sftp://domain.com%5Cuser@host.com:22/some/path/to/file.txt"
	fs := NewFileSystem()
	f, err := fs.NewFile("domain.com%5Cuser@host.com:22", "/some/path/to/file.txt")
	ts.Require().NoError(err)
	ts.Equal(expected, f.URI(), "URI test")
}

func (ts *fileTestSuite) TestStringer() {
	expected := "sftp://user@host.com:22/some/path/to/file.txt"
	ts.Equal(expected, ts.testFile.String(), "String test")
}

func (ts *fileTestSuite) TestNewFile() {
	fs := &FileSystem{}
	// fs is nil
	_, err := fs.NewFile("user@host.com", "")
	ts.Require().Errorf(err, "non-nil sftp.FileSystem pointer is required")

	// authority is ""
	_, err = fs.NewFile("", "asdf")
	ts.Require().Errorf(err, "non-empty strings for bucket and key are required")
	// path is ""
	_, err = fs.NewFile("user@host.com", "")
	ts.Require().Errorf(err, "non-empty strings for bucket and key are required")

	authority := "user@host.com"
	key := "/path/to/file"
	sftpFile, err := fs.NewFile(authority, key)
	ts.Require().NoError(err, "newFile should succeed")
	ts.IsType(&File{}, sftpFile, "newFile returned a File struct")
	ts.Equal(authority, sftpFile.Location().Volume())
	ts.Equal(key, sftpFile.Path())
}

func (ts *fileTestSuite) TestSetDefaultPermissions() {
	testCases := []struct {
		name           string
		client         *mocks.Client
		options        *Options
		expectedError  bool
		expectedErrMsg string
	}{
		{
			name: "No options provided",
			client: func() *mocks.Client {
				client := mocks.NewClient(ts.T())
				return client
			}(),
			options:       nil,
			expectedError: false,
		},
		{
			name: "Default permissions set",
			client: func() *mocks.Client {
				client := mocks.NewClient(ts.T())
				client.EXPECT().Chmod("/some/path.txt", os.FileMode(0o644)).Return(nil)
				return client
			}(),
			options:       &Options{FilePermissions: utils.Ptr("0644")},
			expectedError: false,
		},
		{
			name: "Chmod returns error",
			client: func() *mocks.Client {
				client := mocks.NewClient(ts.T())
				client.EXPECT().Chmod("/some/path.txt", os.FileMode(0o644)).Return(errors.New("chmod error"))
				return client
			}(),
			options:        &Options{FilePermissions: utils.Ptr("0644")},
			expectedError:  true,
			expectedErrMsg: "chmod error",
		},
	}

	for _, tc := range testCases {
		ts.Run(tc.name, func() {
			file := &File{
				path:       "/some/path.txt",
				fileSystem: &FileSystem{options: tc.options},
			}

			err := file.setPermissions(tc.client, tc.options)
			if tc.expectedError {
				ts.Require().Error(err)
				ts.Contains(err.Error(), tc.expectedErrMsg)
			} else {
				ts.Require().NoError(err)
			}
		})
	}
}

func TestFile(t *testing.T) {
	suite.Run(t, &fileTestSuite{})
}
