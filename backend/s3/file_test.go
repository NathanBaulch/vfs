package s3

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/c2fo/vfs/v6"
	"github.com/c2fo/vfs/v6/backend/s3/mocks"
	vfsmocks "github.com/c2fo/vfs/v6/mocks"
	"github.com/c2fo/vfs/v6/options/delete"
	"github.com/c2fo/vfs/v6/options/newfile"
	"github.com/c2fo/vfs/v6/utils"
)

type fileTestSuite struct {
	suite.Suite
}

var (
	s3cliMock      *mocks.Client
	fs             FileSystem
	testFile       vfs.File
	defaultOptions *Options
	testFileName   string
	bucket         string
	matchContext   = mock.MatchedBy(func(context.Context) bool { return true })
)

func (ts *fileTestSuite) SetupTest() {
	var err error
	s3cliMock = mocks.NewClient(ts.T())
	defaultOptions = &Options{AccessKeyID: "abc"}
	fs = FileSystem{client: s3cliMock, options: defaultOptions}
	testFileName = "/some/path/to/file.txt"
	bucket = "bucket"
	testFile, err = fs.NewFile(bucket, testFileName)
	ts.Require().NoError(err, "Shouldn't return error creating test s3.File instance.")
}

func (ts *fileTestSuite) TearDownTest() {
}

func (ts *fileTestSuite) TestRead() {
	contents := "hello world!"

	file, err := fs.NewFile("bucket", "/some/path/file.txt")
	ts.Require().NoError(err, "Shouldn't fail creating new file")

	localFile := &bytes.Buffer{}
	s3cliMock.EXPECT().
		HeadObject(matchContext, mock.AnythingOfType("*s3.HeadObjectInput")).
		Return(&s3.HeadObjectOutput{ContentLength: aws.Int64(12)}, nil).
		Twice()
	s3cliMock.EXPECT().
		GetObject(matchContext, mock.AnythingOfType("*s3.GetObjectInput")).
		Return(&s3.GetObjectOutput{Body: io.NopCloser(strings.NewReader(contents))}, nil).
		Once()
	_, err = io.Copy(localFile, file)
	ts.Require().NoError(err, "no error expected")
	err = file.Close()
	ts.Require().NoError(err, "no error expected")
	ts.Equal(contents, localFile.String(), "Copying an s3 file to a buffer should fill buffer with file's contents")

	// test read with error
	someErr := errors.New("some error")
	s3cliMock.EXPECT().
		GetObject(matchContext, mock.AnythingOfType("*s3.GetObjectInput")).
		Return(nil, someErr).
		Once()
	_, err = io.Copy(localFile, file)
	ts.Require().ErrorIs(err, someErr, "error expected")
	err = file.Close()
	ts.Require().NoError(err, "no error expected")
}

func (ts *fileTestSuite) TestWrite() {
	file, err := fs.NewFile("bucket", "/tmp/hello.txt")
	ts.Require().NoError(err, "Shouldn't fail creating new file")

	contents := []byte("Hello world!")
	count, err := file.Write(contents)
	ts.Require().NoError(err, "Error should be nil when calling Write")
	ts.Len(contents, count, "Returned count of bytes written should match number of bytes passed to Write.")
}

func (ts *fileTestSuite) TestSeek() {
	contents := "hello world!"
	file, err := fs.NewFile("bucket", "/tmp/hello.txt")
	ts.Require().NoError(err, "Shouldn't fail creating new file")

	// setup mock for Size(getHeadObject)
	headOutput := &s3.HeadObjectOutput{ContentLength: aws.Int64(12)}

	testCases := []struct {
		seekOffset  int64
		seekWhence  int
		expectedPos int64
		expectedErr bool
		readContent string
	}{
		{6, 0, 6, false, "world!"},
		{0, 0, 0, false, contents},
		{0, 2, 12, false, ""},
		{-1, 0, 0, true, ""}, // Seek before start
		{0, 3, 0, true, ""},  // bad whence
	}

	for _, tc := range testCases {
		ts.Run(fmt.Sprintf("SeekOffset %d Whence %d", tc.seekOffset, tc.seekWhence), func() {
			m := s3cliMock.EXPECT().
				HeadObject(matchContext, mock.AnythingOfType("*s3.HeadObjectInput")).
				Return(headOutput, nil)
			if tc.expectedErr || tc.seekWhence == io.SeekEnd {
				m.Once()
			} else {
				m.Twice()
			}
			localFile := &bytes.Buffer{}
			pos, err := file.Seek(tc.seekOffset, tc.seekWhence)

			if tc.expectedErr {
				ts.Require().Error(err, "Expected error for seek offset %d and whence %d", tc.seekOffset, tc.seekWhence)
			} else {
				ts.Require().NoError(err, "No error expected for seek offset %d and whence %d", tc.seekOffset, tc.seekWhence)
				ts.Equal(tc.expectedPos, pos, "Expected position does not match for seek offset %d and whence %d", tc.seekOffset, tc.seekWhence)

				// Mock the GetObject call
				m := s3cliMock.EXPECT().GetObject(matchContext, mock.AnythingOfType("*s3.GetObjectInput")).
					Return(&s3.GetObjectOutput{Body: io.NopCloser(strings.NewReader(tc.readContent))}, nil)
				if tc.seekWhence != io.SeekEnd {
					m.Once()
				}

				_, err = io.Copy(localFile, file)
				ts.Require().NoError(err, "No error expected during io.Copy")
				ts.Equal(tc.readContent, localFile.String(), "Content does not match after seek and read")
			}
		})
	}

	// test fails with Size error
	s3cliMock := mocks.NewClient(ts.T())
	fs.client = s3cliMock
	s3cliMock.EXPECT().
		HeadObject(matchContext, mock.AnythingOfType("*s3.HeadObjectInput")).
		Return(nil, &types.NotFound{}).
		Once()
	_, err = file.Seek(0, io.SeekStart)
	ts.Require().ErrorIs(err, vfs.ErrNotExist, "error expected")

	err = file.Close()
	ts.Require().NoError(err, "Closing file should not produce an error")
}

func (ts *fileTestSuite) TestReadEOFSeenReset() {
	contents := "hello world!"
	file, err := fs.NewFile("bucket", "/tmp/hello.txt")
	ts.Require().NoError(err, "Shouldn't fail creating new file")

	s3cliMock.EXPECT().HeadObject(matchContext, mock.AnythingOfType("*s3.HeadObjectInput")).
		Return(&s3.HeadObjectOutput{ContentLength: aws.Int64(int64(len(contents)))}, nil).
		Maybe()

	s3cliMock.EXPECT().GetObject(matchContext, mock.AnythingOfType("*s3.GetObjectInput")).
		Return(&s3.GetObjectOutput{Body: io.NopCloser(strings.NewReader(contents))}, nil).
		Once()

	_, err = io.ReadAll(file)
	ts.Require().NoError(err, "Shouldn't fail reading file")
	ts.True(file.(*File).readEOFSeen, "readEOFSeen should be true after reading the file")

	// Reset cursor to the beginning of the file
	_, err = file.Seek(0, io.SeekStart)
	ts.Require().NoError(err, "Shouldn't fail seeking file")
	ts.False(file.(*File).readEOFSeen, "readEOFSeen should be reset after seeking to the beginning")
}

func (ts *fileTestSuite) TestGetLocation() {
	file, err := fs.NewFile("bucket", "/path/hello.txt")
	ts.Require().NoError(err, "Shouldn't fail creating new file.")

	location := file.Location()
	ts.Equal("s3", location.FileSystem().Scheme(), "Should initialize location with FS underlying file.")
	ts.Equal("/path/", location.Path(), "Should initialize path with the location of the file.")
	ts.Equal("bucket", location.Volume(), "Should initialize bucket with the bucket containing the file.")
}

func (ts *fileTestSuite) TestExists() {
	file, err := fs.NewFile("bucket", "/path/hello.txt")
	ts.Require().NoError(err, "Shouldn't fail creating new file.")

	s3cliMock.EXPECT().HeadObject(matchContext, mock.AnythingOfType("*s3.HeadObjectInput")).Return(&s3.HeadObjectOutput{}, nil)

	exists, err := file.Exists()
	ts.Require().NoError(err, "Shouldn't return an error when exists is true")
	ts.True(exists, "Should return true for exists based on this setup")
}

func (ts *fileTestSuite) TestNotExists() {
	file, err := fs.NewFile("bucket", "/path/hello.txt")
	ts.Require().NoError(err, "Shouldn't fail creating new file.")

	s3cliMock.EXPECT().HeadObject(matchContext, mock.AnythingOfType("*s3.HeadObjectInput")).
		Return(&s3.HeadObjectOutput{}, &types.NotFound{})

	exists, err := file.Exists()
	ts.Require().NoError(err, "Error from key not existing should be hidden since it just confirms it doesn't")
	ts.False(exists, "Should return false for exists based on setup")
}

func (ts *fileTestSuite) TestCopyToFile() {
	targetFile := &File{
		fileSystem: &FileSystem{
			client:  s3cliMock,
			options: defaultOptions,
		},
		bucket: "TestBucket",
		key:    "testKey.txt",
	}

	s3cliMock.EXPECT().CopyObject(matchContext, mock.AnythingOfType("*s3.CopyObjectInput")).Return(&s3.CopyObjectOutput{}, nil)

	err := testFile.CopyToFile(targetFile)
	ts.Require().NoError(err, "Error shouldn't be returned from successful call to CopyToFile")

	// Test With Non Minimum Buffer Size in TouchCopyBuffered
	originalBufferSize := defaultOptions.FileBufferSize
	defaultOptions.FileBufferSize = 2 * utils.TouchCopyMinBufferSize
	targetFile = &File{
		fileSystem: &FileSystem{
			client:  s3cliMock,
			options: defaultOptions,
		},
		bucket: "TestBucket",
		key:    "testKey.txt",
	}
	defaultOptions.FileBufferSize = originalBufferSize

	s3cliMock.EXPECT().CopyObject(matchContext, mock.AnythingOfType("*s3.CopyObjectInput")).Return(&s3.CopyObjectOutput{}, nil)

	err = testFile.CopyToFile(targetFile)
	ts.Require().NoError(err, "Error shouldn't be returned from successful call to CopyToFile")
}

func (ts *fileTestSuite) TestEmptyCopyToFile() {
	targetFile := vfsmocks.NewFile(ts.T())
	targetFile.EXPECT().Write(mock.Anything).Return(0, nil)
	targetFile.EXPECT().Close().Return(nil)
	s3cliMock.EXPECT().
		HeadObject(matchContext, mock.AnythingOfType("*s3.HeadObjectInput")).
		Return(&s3.HeadObjectOutput{ContentLength: aws.Int64(0)}, nil).
		Once()
	err := testFile.CopyToFile(targetFile)
	ts.Require().NoError(err, "Error shouldn't be returned from successful call to CopyToFile")
}

func (ts *fileTestSuite) TestMoveToFile() {
	targetFile := &File{
		fileSystem: &FileSystem{
			client:  s3cliMock,
			options: defaultOptions,
		},
		bucket: "TestBucket",
		key:    "testKey.txt",
	}

	s3cliMock.EXPECT().CopyObject(matchContext, mock.AnythingOfType("*s3.CopyObjectInput")).Return(&s3.CopyObjectOutput{}, nil)
	s3cliMock.EXPECT().DeleteObject(matchContext, mock.AnythingOfType("*s3.DeleteObjectInput")).Return(&s3.DeleteObjectOutput{}, nil)

	err := testFile.MoveToFile(targetFile)
	ts.Require().NoError(err, "Error shouldn't be returned from successful call to MoveToFile")
}

func (ts *fileTestSuite) TestGetCopyObject() {
	testCases := []struct {
		key, expectedCopySource string
	}{
		{
			key:                "/path/to/nospace.txt",
			expectedCopySource: "%2Fpath%2Fto%2Fnospace.txt",
		},
		{
			key:                "/path/to/has space.txt",
			expectedCopySource: "%2Fpath%2Fto%2Fhas%20space.txt",
		},
		{
			key:                "/path/to/encoded%20space.txt",
			expectedCopySource: "%2Fpath%2Fto%2Fencoded%2520space.txt",
		},
		{
			key:                "/path/to/has space/file.txt",
			expectedCopySource: "%2Fpath%2Fto%2Fhas%20space%2Ffile.txt",
		},
		{
			key:                "/path/to/encoded%20space/file.txt",
			expectedCopySource: "%2Fpath%2Fto%2Fencoded%2520space%2Ffile.txt",
		},
	}

	// ensure spaces are properly encoded (or not)
	for _, tc := range testCases {
		ts.Run(tc.key, func() {
			sourceFile := &File{
				fileSystem: &FileSystem{
					client: s3cliMock,
					options: &Options{
						AccessKeyID:                 "abc",
						DisableServerSideEncryption: true,
					},
				},
				bucket: "TestBucket",
				key:    tc.key,
			}

			targetFile := &File{
				fileSystem: &FileSystem{
					client: s3cliMock,
					options: &Options{
						AccessKeyID: "abc",
					},
				},
				bucket: "TestBucket",
				key:    "source.txt",
			}

			// copy from t.key to /source.txt
			actual := sourceFile.getCopyObjectInput(targetFile)
			ts.Equal("TestBucket"+tc.expectedCopySource, *actual.CopySource)
			ts.Empty(actual.ServerSideEncryption, "sse is disabled")
		})
	}

	// test that different options returns nil
	// nil means we can't do s3-to-s3 copy so use TouchCopy
	sourceFile := &File{
		fileSystem: &FileSystem{
			client:  s3cliMock,
			options: defaultOptions,
		},
		bucket: "TestBucket",
		key:    "/path/to/file.txt",
	}

	targetFile := &File{
		fileSystem: &FileSystem{
			client: s3cliMock,
			options: &Options{
				AccessKeyID: "xyz",
				ACL:         "SomeCannedACL",
			},
		},
		bucket: "TestBucket",
		key:    "/path/to/otherFile.txt",
	}
	actual := sourceFile.getCopyObjectInput(targetFile)
	ts.Nil(actual, "copyObjectInput should be nil (can't do s3-to-s3 copyObject)")
}

func (ts *fileTestSuite) TestMoveToFile_CopyError() {
	targetFile := &File{
		fileSystem: &FileSystem{
			client:  s3cliMock,
			options: defaultOptions,
		},
		bucket: "TestBucket",
		key:    "testKey.txt",
	}

	s3cliMock.EXPECT().CopyObject(matchContext, mock.AnythingOfType("*s3.CopyObjectInput")).Return(nil, errors.New("some copy error"))

	err := testFile.MoveToFile(targetFile)
	ts.Require().Error(err, "Error shouldn't be returned from successful call to CopyToFile")
}

func (ts *fileTestSuite) TestCopyToLocation() {
	s3Mock1 := mocks.NewClient(ts.T())
	s3Mock1.EXPECT().CopyObject(matchContext, mock.AnythingOfType("*s3.CopyObjectInput")).Return(nil, nil)
	f := &File{
		fileSystem: &FileSystem{
			client:  s3Mock1,
			options: defaultOptions,
		},
		bucket: "bucket",
		key:    "/hello.txt",
	}

	defer func() {
		err := f.Close()
		ts.Require().NoError(err, "no error expected")
	}()

	l := &Location{
		fileSystem: &FileSystem{
			client:  mocks.NewClient(ts.T()),
			options: defaultOptions,
		},
		bucket: "bucket",
		prefix: "/subdir/",
	}

	// no error "copying" objects
	_, err := f.CopyToLocation(l)
	ts.Require().NoError(err, "Shouldn't return error for this call to CopyToLocation")
}

func (ts *fileTestSuite) TestTouch() {
	// Copy portion tested through CopyToLocation, just need to test whether Delete happens
	// in addition to CopyToLocation

	s3Mock1 := mocks.NewClient(ts.T())
	s3Mock1.EXPECT().HeadObject(matchContext, mock.AnythingOfType("*s3.HeadObjectInput")).Return(&s3.HeadObjectOutput{}, nil)
	s3Mock1.EXPECT().CopyObject(matchContext, mock.AnythingOfType("*s3.CopyObjectInput")).Return(nil, nil)
	s3Mock1.EXPECT().DeleteObject(matchContext, mock.AnythingOfType("*s3.DeleteObjectInput")).Return(&s3.DeleteObjectOutput{}, nil)

	file := &File{
		fileSystem: &FileSystem{
			client:  s3Mock1,
			options: defaultOptions,
		},
		bucket: "newBucket",
		key:    "/new/file/path/hello.txt",
	}

	err := file.Touch()
	ts.Require().NoError(err, "Shouldn't return error creating test s3.File instance.")

	// test non-existent length
	s3Mock2 := mocks.NewClient(ts.T())
	s3Mock2.EXPECT().HeadObject(matchContext, mock.AnythingOfType("*s3.HeadObjectInput")).
		Return(&s3.HeadObjectOutput{}, &types.NotFound{}).Once()
	s3Mock2.EXPECT().HeadObject(matchContext, mock.AnythingOfType("*s3.HeadObjectInput")).
		Return(&s3.HeadObjectOutput{}, nil)
	file2 := &File{
		fileSystem: &FileSystem{
			client:  s3Mock2,
			options: defaultOptions,
		},
		bucket: "newBucket",
		key:    "/new/file/path/hello.txt",
	}

	s3Mock2.EXPECT().PutObject(matchContext, mock.AnythingOfType("*s3.PutObjectInput"), mock.Anything, mock.Anything).
		Return(&s3.PutObjectOutput{}, nil)

	terr2 := file2.Touch()
	ts.Require().NoError(terr2, "Shouldn't return error creating test s3.File instance.")
}

func (ts *fileTestSuite) TestMoveToLocation() {
	// Copy portion tested through CopyToLocation, just need to test whether Delete happens
	// in addition to CopyToLocation
	s3Mock1 := mocks.NewClient(ts.T())
	f := &File{
		fileSystem: &FileSystem{
			client:  s3Mock1,
			options: defaultOptions,
		},
		bucket: "newBucket",
		key:    "/new/file/path/hello.txt",
	}
	location := vfsmocks.NewLocation(ts.T())
	location.EXPECT().NewFile(mock.Anything).Return(f, nil)

	s3cliMock.EXPECT().CopyObject(matchContext, mock.AnythingOfType("*s3.CopyObjectInput")).Return(&s3.CopyObjectOutput{}, nil)
	s3cliMock.EXPECT().DeleteObject(matchContext, mock.AnythingOfType("*s3.DeleteObjectInput")).Return(&s3.DeleteObjectOutput{}, nil)

	file, err := fs.NewFile("bucket", "/hello.txt")
	ts.Require().NoError(err, "Shouldn't return error creating test s3.File instance.")

	defer func() {
		err := file.Close()
		ts.Require().NoError(err, "no error expected")
	}()

	_, err = file.MoveToLocation(location)
	ts.Require().NoError(err, "no error expected")

	// test non-scheme MoveToLocation
	mockLocation := vfsmocks.NewLocation(ts.T())
	mockLocation.EXPECT().NewFile(mock.Anything).
		Return(&File{fileSystem: &FileSystem{client: s3Mock1}, bucket: "bucket", key: "/new/hello.txt"}, nil)

	s3cliMock2 := mocks.NewClient(ts.T())
	s3cliMock2.EXPECT().CopyObject(matchContext, mock.AnythingOfType("*s3.CopyObjectInput")).Return(&s3.CopyObjectOutput{}, nil)

	fs = FileSystem{client: s3cliMock2}
	file2, err := fs.NewFile("bucket", "/hello.txt")
	ts.Require().NoError(err, "Shouldn't return error creating test s3.File instance.")

	_, err = file2.CopyToLocation(mockLocation)
	ts.Require().NoError(err, "MoveToLocation error not expected")
}

func (ts *fileTestSuite) TestMoveToLocationFail() {
	// If CopyToLocation fails we need to ensure DeleteObject isn't called.
	location := vfsmocks.NewLocation(ts.T())
	location.EXPECT().NewFile(mock.Anything).Return(&File{fileSystem: &fs, bucket: "bucket", key: "/new/hello.txt"}, nil)

	s3cliMock.EXPECT().CopyObject(matchContext, mock.AnythingOfType("*s3.CopyObjectInput")).Return(nil, errors.New("didn't copy, oh noes"))

	file, err := fs.NewFile("bucket", "/hello.txt")
	ts.Require().NoError(err, "Shouldn't return error creating test s3.File instance.")

	_, err = file.MoveToLocation(location)
	ts.Require().Error(err, "MoveToLocation error not expected")

	err = file.Close()
	ts.Require().NoError(err, "no close error expected")
}

func (ts *fileTestSuite) TestDelete() {
	s3cliMock.EXPECT().DeleteObject(matchContext, mock.AnythingOfType("*s3.DeleteObjectInput")).Return(&s3.DeleteObjectOutput{}, nil)
	err := testFile.Delete()
	ts.Require().NoError(err, "Successful delete should not return an error.")
}

func (ts *fileTestSuite) TestDeleteError() {
	s3cliMock.EXPECT().DeleteObject(matchContext, mock.AnythingOfType("*s3.DeleteObjectInput")).Return(nil, errors.New("something went wrong"))
	err := testFile.Delete()
	ts.Require().EqualError(err, "something went wrong", "Delete should return an error if s3 api had error.")
}

func (ts *fileTestSuite) TestDeleteWithAllVersionsOption() {
	versOutput := s3.ListObjectVersionsOutput{
		Versions: []types.ObjectVersion{
			{VersionId: aws.String("ver1")},
			{VersionId: aws.String("ver2")},
		},
	}
	s3cliMock.EXPECT().ListObjectVersions(matchContext, mock.AnythingOfType("*s3.ListObjectVersionsInput")).Return(&versOutput, nil)
	s3cliMock.EXPECT().DeleteObject(matchContext, mock.AnythingOfType("*s3.DeleteObjectInput")).Return(&s3.DeleteObjectOutput{}, nil).Times(3)

	err := testFile.Delete(delete.WithAllVersions())
	ts.Require().NoError(err, "Successful delete should not return an error.")
}

func (ts *fileTestSuite) TestDeleteWithAllVersionsOptionError() {
	versOutput := s3.ListObjectVersionsOutput{
		Versions: []types.ObjectVersion{
			{VersionId: aws.String("ver1")},
			{VersionId: aws.String("ver2")},
		},
	}
	s3cliMock.EXPECT().ListObjectVersions(matchContext, mock.AnythingOfType("*s3.ListObjectVersionsInput")).
		Return(&versOutput, nil)
	key := utils.Ptr(utils.RemoveLeadingSlash(testFileName))
	s3cliMock.EXPECT().DeleteObject(matchContext, &s3.DeleteObjectInput{Key: key, Bucket: &bucket}).
		Return(&s3.DeleteObjectOutput{}, nil).
		Once()
	s3cliMock.EXPECT().DeleteObject(matchContext, &s3.DeleteObjectInput{Key: key, Bucket: &bucket, VersionId: aws.String("ver1")}).
		Return(nil, errors.New("something went wrong")).
		Once()

	err := testFile.Delete(delete.WithAllVersions())
	ts.Require().Error(err, "Delete should return an error if s3 api had error.")
}

func (ts *fileTestSuite) TestLastModified() {
	now := time.Now()
	s3cliMock.EXPECT().HeadObject(matchContext, mock.AnythingOfType("*s3.HeadObjectInput")).Return(&s3.HeadObjectOutput{
		LastModified: &now,
	}, nil)
	modTime, err := testFile.LastModified()
	ts.Require().NoError(err, "Error should be nil when correctly returning time of object.")
	ts.Equal(&now, modTime, "Returned time matches expected LastModified time.")
}

func (ts *fileTestSuite) TestLastModifiedFail() {
	// setup error on HEAD
	s3cliMock.EXPECT().HeadObject(matchContext, mock.AnythingOfType("*s3.HeadObjectInput")).Return(nil,
		errors.New("boom"))
	m, e := testFile.LastModified()
	ts.Require().Error(e, "got error as expected")
	ts.Nil(m, "nil ModTime returned")
}

func (ts *fileTestSuite) TestName() {
	ts.Equal("file.txt", testFile.Name(), "Name should return just the name of the file.")
}

func (ts *fileTestSuite) TestSize() {
	contentLength := int64(100)
	s3cliMock.EXPECT().HeadObject(matchContext, mock.AnythingOfType("*s3.HeadObjectInput")).Return(&s3.HeadObjectOutput{
		ContentLength: &contentLength,
	}, nil)

	size, err := testFile.Size()
	ts.Require().NoError(err, "Error should be nil when requesting size for file that exists.")
	ts.Equal(uint64(100), size, "Size should return the ContentLength value from s3 HEAD request.")
}

func (ts *fileTestSuite) TestPath() {
	ts.Equal("/some/path/to/file.txt", testFile.Path(), "Should return file.key (with leading slash)")
}

func (ts *fileTestSuite) TestURI() {
	s3cliMock = mocks.NewClient(ts.T())
	fs = FileSystem{client: s3cliMock}
	file, _ := fs.NewFile("mybucket", "/some/file/test.txt")
	expected := "s3://mybucket/some/file/test.txt"
	ts.Equal(expected, file.URI(), "%s does not match %s", file.URI(), expected)
}

func (ts *fileTestSuite) TestStringer() {
	fs = FileSystem{client: mocks.NewClient(ts.T())}
	file, _ := fs.NewFile("mybucket", "/some/file/test.txt")
	ts.Equal("s3://mybucket/some/file/test.txt", file.String())
}

func (ts *fileTestSuite) TestUploadInput() {
	fs = FileSystem{client: mocks.NewClient(ts.T())}
	file, _ := fs.NewFile("mybucket", "/some/file/test.txt")
	ts.Equal(types.ServerSideEncryptionAes256, uploadInput(file.(*File)).ServerSideEncryption, "sse was set")
	ts.Equal("some/file/test.txt", *uploadInput(file.(*File)).Key, "key was set")
	ts.Equal("mybucket", *uploadInput(file.(*File)).Bucket, "bucket was set")
}

func (ts *fileTestSuite) TestUploadInputDisableSSE() {
	fs := NewFileSystem().
		WithOptions(Options{DisableServerSideEncryption: true})
	file, _ := fs.NewFile("mybucket", "/some/file/test.txt")
	input := uploadInput(file.(*File))
	ts.Empty(input.ServerSideEncryption, "sse was disabled")
	ts.Equal("some/file/test.txt", *input.Key, "key was set")
	ts.Equal("mybucket", *input.Bucket, "bucket was set")
}

func (ts *fileTestSuite) TestUploadInputContentType() {
	fs = FileSystem{client: mocks.NewClient(ts.T())}
	file, _ := fs.NewFile("mybucket", "/some/file/test.txt", newfile.WithContentType("text/plain"))
	input := uploadInput(file.(*File))
	ts.Equal("text/plain", *input.ContentType)
}

func (ts *fileTestSuite) TestNewFile() {
	fs := &FileSystem{}
	// fs is nil
	_, err := fs.NewFile("", "")
	ts.Require().Errorf(err, "non-nil s3.FileSystem pointer is required")

	// bucket is ""
	_, err = fs.NewFile("", "asdf")
	ts.Require().Errorf(err, "non-empty strings for bucket and key are required")
	// key is ""
	_, err = fs.NewFile("asdf", "")
	ts.Require().Errorf(err, "non-empty strings for bucket and key are required")

	//
	bucket := "mybucket"
	key := "/path/to/key"
	file, err := fs.NewFile(bucket, key)
	ts.Require().NoError(err, "newFile should succeed")
	ts.IsType(&File{}, file, "newFile returned a File struct")
	ts.Equal(bucket, file.Location().Volume())
	ts.Equal(key, file.Path())
}

func (ts *fileTestSuite) TestCloseWithoutWrite() {
	fs := &FileSystem{}
	file, err := fs.NewFile("mybucket", "/some/file/test.txt")
	ts.Require().NoError(err)
	ts.Require().NoError(file.Close())
}

func (ts *fileTestSuite) TestCloseWithWrite() {
	s3Mock := mocks.NewClient(ts.T())
	s3Mock.EXPECT().HeadObject(matchContext, mock.AnythingOfType("*s3.HeadObjectInput")).
		Return(&s3.HeadObjectOutput{}, &types.NotFound{})
	s3Mock.EXPECT().PutObject(matchContext, mock.AnythingOfType("*s3.PutObjectInput"), mock.Anything, mock.Anything).
		Return(&s3.PutObjectOutput{}, nil)
	file := &File{
		fileSystem: &FileSystem{
			client:  s3Mock,
			options: defaultOptions,
		},
		bucket: "newBucket",
		key:    "/new/file/path/hello.txt",
	}
	contents := []byte("Hello world!")
	_, err := file.Write(contents)
	ts.Require().NoError(err, "Error should be nil when calling Write")
	err = file.Close()
	ts.Require().Error(err, "file doesn't exists, retired 5 times")
}

func (ts *fileTestSuite) TestWriteOperations() {
	var contents *string
	setup := func(s3Mock *mocks.Client) {
		s3Mock.EXPECT().PutObject(matchContext, mock.AnythingOfType("*s3.PutObjectInput"), mock.Anything, mock.Anything).
			RunAndReturn(func(_ context.Context, input *s3.PutObjectInput, _ ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
				// Read from the input.Body (which is a PipeReader) to simulate actual upload
				b, err := io.ReadAll(input.Body)
				if err != nil {
					panic(err)
				}
				contents = utils.Ptr(string(b))
				return &s3.PutObjectOutput{}, nil
			})
	}

	testCases := []struct {
		name             string
		setup            func(*mocks.Client) *File // Function to set up each test case
		actions          []func(*File) error       // Actions to perform on the file (Write, Seek, etc.)
		wantErr          bool
		validate         func(*File) error // Additional validations if needed
		expectedContents string
	}{
		{
			name: "Write and Close - Close failure",
			setup: func(s3Mock *mocks.Client) *File {
				// Mock setup specific to this test case
				s3Mock.EXPECT().HeadObject(matchContext, mock.AnythingOfType("*s3.HeadObjectInput")).
					Return(&s3.HeadObjectOutput{}, &types.NotFound{}).Times(5)
				// Return a new File instance with this specific mock configuration
				return &File{
					fileSystem: &FileSystem{
						client:  s3Mock,
						options: defaultOptions,
					},
					bucket: "newBucket",
					key:    "/new/file/path/hello.txt",
				}
			},
			actions: []func(*File) error{
				func(f *File) error {
					_, err := f.Write([]byte("Hello world!"))
					return err
				},
				func(f *File) error {
					return f.Close()
				},
			},
			wantErr: true,
		},
		{
			name: "Write and Close - success",
			setup: func(s3Mock *mocks.Client) *File {
				// Mock setup specific to this test case
				s3Mock.EXPECT().HeadObject(matchContext, mock.AnythingOfType("*s3.HeadObjectInput")).
					Return(&s3.HeadObjectOutput{}, nil).Once()
				// Return a new File instance with this specific mock configuration
				return &File{
					fileSystem: &FileSystem{
						client:  s3Mock,
						options: defaultOptions,
					},
					bucket: "newBucket",
					key:    "/new/file/path/hello.txt",
				}
			},
			actions: []func(*File) error{
				func(f *File) error {
					_, err := f.Write([]byte("Hello world!"))
					return err
				},
				func(f *File) error {
					return f.Close()
				},
			},
			wantErr:          false,
			expectedContents: `Hello world!`,
		},
		{
			name: "Write, Seek, Write and Close new file - success",
			setup: func(s3Mock *mocks.Client) *File {
				// Mock setup specific to this test case
				s3Mock.EXPECT().HeadObject(matchContext, mock.AnythingOfType("*s3.HeadObjectInput")).
					Return(nil, &types.NotFound{}).Twice()
				s3Mock.EXPECT().HeadObject(matchContext, mock.AnythingOfType("*s3.HeadObjectInput")).
					Return(&s3.HeadObjectOutput{}, nil).Once()

				// Return a new File instance with this specific mock configuration
				return &File{
					fileSystem: &FileSystem{
						client:  s3Mock,
						options: defaultOptions,
					},
					bucket: "newBucket",
					key:    "/new/file/path/hello.txt",
				}
			},
			actions: []func(*File) error{
				func(f *File) error {
					_, err := f.Write([]byte("Hello world!"))
					return err
				},
				func(f *File) error {
					_, err := f.Seek(6, io.SeekStart)
					return err
				},
				func(f *File) error {
					_, err := f.Write([]byte("Bob!"))
					return err
				},
				func(*File) error {
					// sleep 1 sec
					time.Sleep(time.Second)
					return nil
				},
				func(f *File) error {
					return f.Close()
				},
			},
			wantErr:          false,
			expectedContents: `Hello Bob!d!`,
		},
	}

	for _, tc := range testCases {
		ts.Run(tc.name, func() {
			contents = nil // reset contents

			s3Mock := mocks.NewClient(ts.T()) // Create a new mock for each test
			setup(s3Mock)
			file := tc.setup(s3Mock) // Set up the file for this test

			var err error
			for _, action := range tc.actions {
				err = action(file)
				if err != nil {
					break
				}
			}

			if tc.wantErr {
				ts.Require().Error(err)
			} else {
				ts.Require().NoError(err)
				ts.Equal(tc.expectedContents, *contents, "Contents of file should match expected contents")
			}

			// TODO: is this even needed?
			if tc.validate != nil {
				err := tc.validate(file)
				ts.Require().NoError(err)
			}
		})
	}
}

func TestFile(t *testing.T) {
	suite.Run(t, &fileTestSuite{})
}
