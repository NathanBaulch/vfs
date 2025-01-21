package gs

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"testing"

	"cloud.google.com/go/storage"
	"github.com/fsouza/fake-gcs-server/fakestorage"
	"github.com/stretchr/testify/suite"

	"github.com/c2fo/vfs/v6/options/delete"
	"github.com/c2fo/vfs/v6/options/newfile"
	"github.com/c2fo/vfs/v6/utils"
)

type fileTestSuite struct {
	suite.Suite
}

func (ts *fileTestSuite) assertFileExists(fs *FileSystem, bucket *storage.BucketHandle, name string, content []byte) {
	objectHandle := bucket.Object(name)
	_, err := objectHandle.Attrs(context.Background())
	ts.Require().NoError(err)

	reader, err := objectHandle.NewReader(context.Background())
	ts.Require().NoError(err)
	defer func() {
		err := reader.Close()
		ts.Require().NoError(err)
	}()
	data, err := io.ReadAll(reader)
	ts.Require().NoError(err)
	ts.Equal(content, data)

	file, err := fs.NewFile(bucket.BucketName(), "/"+name)
	ts.Require().NoError(err)
	exists, err := file.Exists()
	ts.Require().NoError(err)
	ts.Require().True(exists)

	data, err = io.ReadAll(file)
	ts.Require().NoError(err)
	ts.Equal(content, data)
}

func (ts *fileTestSuite) assertFileNotExists(fs *FileSystem, bucket *storage.BucketHandle, name string) {
	objectHandle := bucket.Object(name)
	_, err := objectHandle.Attrs(context.Background())
	ts.Require().ErrorIs(err, storage.ErrObjectNotExist)

	file, err := fs.NewFile(bucket.BucketName(), "/"+name)
	ts.Require().NoError(err)
	exists, err := file.Exists()
	ts.Require().NoError(err)
	ts.Require().False(exists)
}

func (ts *fileTestSuite) TestRead() {
	contents := "hello world!"
	bucketName := "bucki"
	objectName := "some/path/file.txt"
	server := fakestorage.NewServer(
		[]fakestorage.Object{
			{
				ObjectAttrs: fakestorage.ObjectAttrs{
					BucketName:      bucketName,
					Name:            objectName,
					ContentType:     "text/plain",
					ContentEncoding: "utf8",
				},
				Content: []byte(contents),
			},
		},
	)
	defer server.Stop()
	fs := NewFileSystem().WithClient(server.Client())

	file, err := fs.NewFile(bucketName, "/"+objectName)
	ts.Require().NoError(err, "Shouldn't fail creating new file")

	localFile := &bytes.Buffer{}

	buffer := make([]byte, utils.TouchCopyMinBufferSize)
	_, err = io.CopyBuffer(localFile, file, buffer)
	ts.Require().NoError(err, "no error expected")
	err = file.Close()
	ts.Require().NoError(err, "no error expected")

	ts.Equal(localFile.String(), contents, "Copying an gs file to a buffer should fill buffer with file's contents")
}

func (ts *fileTestSuite) TestDelete() {
	contents := "hello world!"
	bucketName := "bucki"
	objectName := "some/path/file.txt"
	server := fakestorage.NewServer(
		[]fakestorage.Object{
			{
				ObjectAttrs: fakestorage.ObjectAttrs{
					BucketName:      bucketName,
					Name:            objectName,
					ContentType:     "text/plain",
					ContentEncoding: "utf8",
				},
				Content: []byte(contents),
			},
		},
	)
	defer server.Stop()
	client := server.Client()
	fs := NewFileSystem().WithClient(client)

	file, err := fs.NewFile(bucketName, "/"+objectName)
	ts.Require().NoError(err, "Shouldn't fail creating new file")

	err = file.Delete()
	ts.Require().NoError(err, "Shouldn't fail deleting the file")

	bucket := client.Bucket(bucketName)
	ts.assertFileNotExists(fs, bucket, objectName)
}

func (ts *fileTestSuite) TestDeleteError() {
	contents := "hello world!"
	bucketName := "bucki"
	objectName := "some/path/file.txt"
	server := fakestorage.NewServer(
		[]fakestorage.Object{
			{
				ObjectAttrs: fakestorage.ObjectAttrs{
					BucketName:      bucketName,
					Name:            objectName,
					ContentType:     "text/plain",
					ContentEncoding: "utf8",
				},
				Content: []byte(contents),
			},
		},
	)
	defer server.Stop()
	client := server.Client()
	fs := NewFileSystem().WithClient(client)

	file, err := fs.NewFile(bucketName, "/invalidObject")
	ts.Require().NoError(err, "Shouldn't fail creating new file")

	err = file.Delete()
	ts.Require().Error(err, "Should return an error if gs client had error")
}

func (ts *fileTestSuite) TestDeleteRemoveAllVersions() {
	contents := "hello world!"
	bucketName := "bucki"
	objectName := "some/path/file.txt"
	server := fakestorage.NewServer(
		[]fakestorage.Object{
			{
				ObjectAttrs: fakestorage.ObjectAttrs{
					BucketName:      bucketName,
					Name:            objectName,
					ContentType:     "text/plain",
					ContentEncoding: "utf8",
				},
				Content: []byte(contents),
			},
		},
	)
	defer server.Stop()
	client := server.Client()
	fs := NewFileSystem().WithClient(client)

	file, err := fs.NewFile(bucketName, "/"+objectName)
	ts.Require().NoError(err, "Shouldn't fail creating new file")

	f := file.(*File)
	handles, err := f.getObjectGenerationHandles()
	ts.Require().NoError(err, "Shouldn't fail getting object generation handles")
	ts.Len(handles, 1)

	err = file.Delete(delete.WithAllVersions())
	ts.Require().NoError(err, "Shouldn't fail deleting the file")

	bucket := client.Bucket(bucketName)
	ts.assertFileNotExists(fs, bucket, objectName)
	handles, err = f.getObjectGenerationHandles()
	ts.Require().NoError(err, "Shouldn't fail getting object generation handles")
	ts.Nil(handles)
}

func (ts *fileTestSuite) TestWrite() {
	contents := "hello world!"
	bucketName := "bucki"
	objectName := "some/path/file.txt"
	server := fakestorage.NewServer([]fakestorage.Object{})
	defer server.Stop()
	fs := NewFileSystem().WithClient(server.Client())

	file, err := fs.NewFile(bucketName, "/"+objectName)
	ts.Require().NoError(err, "Shouldn't fail creating new file")

	count, err := file.Write([]byte(contents))
	ts.Require().NoError(err, "Error should be nil when calling Write")
	ts.Len(contents, count, "Returned count of bytes written should match number of bytes passed to Write.")
}

func (ts *fileTestSuite) TestWriteWithContentType() {
	contents := "hello world!"
	bucketName := "bucki"
	objectName := "some/path/file.txt"
	server := fakestorage.NewServer([]fakestorage.Object{})
	defer server.Stop()
	client := server.Client()
	bucket := client.Bucket(bucketName)
	ctx := context.Background()
	err := bucket.Create(ctx, "", nil)
	ts.Require().NoError(err)
	fs := NewFileSystem().WithClient(client)

	file, err := fs.NewFile(bucketName, "/"+objectName, newfile.WithContentType("text/plain"))
	ts.Require().NoError(err, "Shouldn't fail creating new file")

	_, err = file.Write([]byte(contents))
	ts.Require().NoError(err, "Error should be nil when calling Write")

	err = file.Close()
	ts.Require().NoError(err, "Error should be nil when calling Close")

	attrs, err := bucket.Object(objectName).Attrs(ctx)
	ts.Require().NoError(err)
	ts.Equal("text/plain", attrs.ContentType)
}

func (ts *fileTestSuite) TestTouchWithContentType() {
	bucketName := "bucki"
	objectName := "some/path/file.txt"
	server := fakestorage.NewServer([]fakestorage.Object{})
	defer server.Stop()
	client := server.Client()
	bucket := client.Bucket(bucketName)
	ctx := context.Background()
	err := bucket.Create(ctx, "", nil)
	ts.Require().NoError(err)
	fs := NewFileSystem().WithClient(client)

	file, err := fs.NewFile(bucketName, "/"+objectName, newfile.WithContentType("text/plain"))
	ts.Require().NoError(err, "Shouldn't fail creating new file")

	err = file.Touch()
	ts.Require().NoError(err, "Error should be nil when calling Touch")

	attrs, err := bucket.Object(objectName).Attrs(ctx)
	ts.Require().NoError(err)
	ts.Equal("text/plain", attrs.ContentType)
}

func (ts *fileTestSuite) TestGetLocation() {
	server := fakestorage.NewServer([]fakestorage.Object{})
	defer server.Stop()
	fs := NewFileSystem().WithClient(server.Client())

	file, err := fs.NewFile("bucket", "/path/hello.txt")
	ts.Require().NoError(err, "Shouldn't fail creating new file.")

	location := file.Location()
	ts.Equal("gs", location.FileSystem().Scheme(), "Should initialize location with FS underlying file.")
	ts.Equal("/path/", location.Path(), "Should initialize path with the location of the file.")
	ts.Equal("bucket", location.Volume(), "Should initialize bucket with the bucket containing the file.")
}

func (ts *fileTestSuite) TestExists() {
	bucketName := "bucki"
	objectName := "some/path/file.txt"

	server := fakestorage.NewServer([]fakestorage.Object{
		{
			ObjectAttrs: fakestorage.ObjectAttrs{
				BucketName:      bucketName,
				Name:            objectName,
				ContentType:     "text/plain",
				ContentEncoding: "utf8",
			},
			Content: []byte("content"),
		},
	})
	defer server.Stop()
	fs := NewFileSystem().WithClient(server.Client())

	file, err := fs.NewFile(bucketName, "/"+objectName)
	ts.Require().NoError(err, "Shouldn't fail creating new file.")

	exists, err := file.Exists()
	ts.Require().NoError(err, "Shouldn't return an error when exists is true")
	ts.True(exists, "Should return true for exists based on this setup")
}

func (ts *fileTestSuite) TestNotExists() {
	server := fakestorage.NewServer([]fakestorage.Object{})
	defer server.Stop()
	fs := NewFileSystem().WithClient(server.Client())

	file, err := fs.NewFile("bucket", "/path/hello.txt")
	ts.Require().NoError(err, "Shouldn't fail creating new file.")

	exists, err := file.Exists()
	ts.Require().NoError(err, "Error from key not existing should be hidden since it just confirms it doesn't")
	ts.False(exists, "Should return false for exists based on setup")
}

func (ts *fileTestSuite) TestMoveAndCopy() {
	type testCase struct {
		move       bool
		readFirst  bool
		sameBucket bool
	}
	var testCases []testCase

	for idx := 0; idx <= (1<<3)-1; idx++ {
		testCases = append(testCases, testCase{
			move:       (idx & (1 << 0)) != 0,
			readFirst:  (idx & (1 << 1)) != 0,
			sameBucket: (idx & (1 << 2)) != 0,
		})
	}

	for _, tc := range testCases {
		ts.Run(fmt.Sprintf("%#v", tc), func() {
			sourceName := "source.txt"
			targetName := "target.txt"
			sourceBucketName := "bucket-source"
			var targetBucketName string
			if tc.sameBucket {
				targetBucketName = sourceBucketName
			} else {
				targetBucketName = "bucket-target"
			}

			content := []byte("content")
			fakeObjects := []fakestorage.Object{
				{
					ObjectAttrs: fakestorage.ObjectAttrs{
						BucketName:      sourceBucketName,
						Name:            sourceName,
						ContentType:     "text/plain",
						ContentEncoding: "utf8",
					},
					Content: content,
				},
			}
			fakeObjects = append(fakeObjects, fakestorage.Object{
				ObjectAttrs: fakestorage.ObjectAttrs{
					BucketName:      targetBucketName,
					Name:            "place.holder",
					ContentType:     "text/plain",
					ContentEncoding: "utf8",
				},
				Content: []byte{},
			})
			server := fakestorage.NewServer(fakeObjects)
			defer server.Stop()
			client := server.Client()
			fs := NewFileSystem().WithClient(client)
			sourceBucket := client.Bucket(sourceBucketName)
			targetBucket := client.Bucket(targetBucketName)

			ts.assertFileExists(fs, sourceBucket, sourceName, content)
			ts.assertFileNotExists(fs, targetBucket, targetName)

			sourceFile, err := fs.NewFile(sourceBucketName, "/"+sourceName)
			ts.Require().NoError(err)
			targetFile, err := fs.NewFile(targetBucketName, "/"+targetName)
			ts.Require().NoError(err)

			if tc.readFirst {
				_, err := io.ReadAll(sourceFile)
				ts.Require().NoError(err)
			}

			if tc.move {
				err = sourceFile.MoveToFile(targetFile)
			} else {
				err = sourceFile.CopyToFile(targetFile)
			}

			if tc.readFirst {
				ts.Require().Error(err, "Error should be returned for operation on file that has been read (i.e. has non 0 cursor position)")
			} else {
				ts.Require().NoError(err, "Error shouldn't be returned from successful operation")

				if tc.move {
					ts.assertFileNotExists(fs, sourceBucket, sourceName)
				} else {
					ts.assertFileExists(fs, sourceBucket, sourceName, content)
				}

				ts.assertFileExists(fs, targetBucket, targetName, content)
			}
		})
	}
}

func (ts *fileTestSuite) TestMoveAndCopyBuffered() {
	type testCase struct {
		move       bool
		readFirst  bool
		sameBucket bool
	}
	var testCases []testCase

	for idx := 0; idx <= (1<<3)-1; idx++ {
		testCases = append(testCases, testCase{
			move:       (idx & (1 << 0)) != 0,
			readFirst:  (idx & (1 << 1)) != 0,
			sameBucket: (idx & (1 << 2)) != 0,
		})
	}

	for _, tc := range testCases {
		ts.Run(fmt.Sprintf("%#v", tc), func() {
			sourceName := "source.txt"
			targetName := "target.txt"
			sourceBucketName := "bucket-source"
			var targetBucketName string
			if tc.sameBucket {
				targetBucketName = sourceBucketName
			} else {
				targetBucketName = "bucket-target"
			}

			content := []byte("content")
			fakeObjects := []fakestorage.Object{
				{
					ObjectAttrs: fakestorage.ObjectAttrs{
						BucketName:      sourceBucketName,
						Name:            sourceName,
						ContentType:     "text/plain",
						ContentEncoding: "utf8",
					},
					Content: content,
				},
			}
			fakeObjects = append(fakeObjects, fakestorage.Object{
				ObjectAttrs: fakestorage.ObjectAttrs{
					BucketName:      targetBucketName,
					Name:            "place.holder",
					ContentType:     "text/plain",
					ContentEncoding: "utf8",
				},
				Content: []byte{},
			})
			server := fakestorage.NewServer(fakeObjects)
			defer server.Stop()
			client := server.Client()
			opts := Options{FileBufferSize: 2 * utils.TouchCopyMinBufferSize}
			fs := NewFileSystem().WithOptions(opts).WithClient(client)
			sourceBucket := client.Bucket(sourceBucketName)
			targetBucket := client.Bucket(targetBucketName)

			ts.assertFileExists(fs, sourceBucket, sourceName, content)
			ts.assertFileNotExists(fs, targetBucket, targetName)

			sourceFile, err := fs.NewFile(sourceBucketName, "/"+sourceName)
			ts.Require().NoError(err)
			targetFile, err := fs.NewFile(targetBucketName, "/"+targetName)
			ts.Require().NoError(err)

			if tc.readFirst {
				_, err := io.ReadAll(sourceFile)
				ts.Require().NoError(err)
			}

			if tc.move {
				err = sourceFile.MoveToFile(targetFile)
			} else {
				err = sourceFile.CopyToFile(targetFile)
			}

			if tc.readFirst {
				ts.Require().Error(err, "Error should be returned for operation on file that has been read (i.e. has non 0 cursor position)")
			} else {
				ts.Require().NoError(err, "Error shouldn't be returned from successful operation")

				if tc.move {
					ts.assertFileNotExists(fs, sourceBucket, sourceName)
				} else {
					ts.assertFileExists(fs, sourceBucket, sourceName, content)
				}

				ts.assertFileExists(fs, targetBucket, targetName, content)
			}
		})
	}
}

func TestFile(t *testing.T) {
	suite.Run(t, &fileTestSuite{})
}
