package ftp

import (
	"errors"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/jlaffaye/ftp"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/c2fo/vfs/v6/backend/ftp/mocks"
	"github.com/c2fo/vfs/v6/backend/ftp/types"
)

type dataConnSuite struct {
	suite.Suite
	ftpFile *File
	client  *mocks.Client
}

func TestDataConn(t *testing.T) {
	suite.Run(t, &dataConnSuite{})
}

// test setup
func (s *dataConnSuite) SetupTest() {
	// set up ftpfile
	filepath := "/some/path.txt"
	s.client = mocks.NewClient(s.T())
	s.ftpFile = &File{
		fileSystem: NewFileSystem().WithClient(s.client),
		path:       filepath,
	}
}

func (s *dataConnSuite) TestGetDataConn_OpenForRead() {
	// dataconn is nil - open for read
	s.client.EXPECT().
		RetrFrom(s.ftpFile.Path(), uint64(0)).
		Return(&ftp.Response{}, nil).
		Once()
	dc, err := getDataConn(s.client, s.ftpFile, types.OpenRead)
	s.Require().NoError(err, "no error expected")
	s.IsTypef(&dataConn{}, dc, "dataconn returned")
}

func (s *dataConnSuite) TestGetDataConn_ReadError() {
	// dataconn is nil - error calling client.RetrFrom
	someErr := errors.New("some error")

	s.client.EXPECT().
		RetrFrom(s.ftpFile.Path(), uint64(0)).
		Return(nil, someErr).
		Once()
	dc, err := getDataConn(s.client, s.ftpFile, types.OpenRead)
	s.Require().ErrorIs(err, someErr, "error is right kind of error")
	s.Nil(dc, "dataconn should be nil on error")
}

func (s *dataConnSuite) TestGetDataConn_WriteLocationNotExists() {
	// dataconn is nil - open for write - location doesn't exist - success
	s.client.EXPECT().
		List("/").
		Return(nil, errors.New("550")).
		Once()
	s.client.EXPECT().
		MakeDir(s.ftpFile.Location().Path()).
		Return(nil).
		Once()
	s.client.EXPECT().
		StorFrom(s.ftpFile.Path(), mock.Anything, uint64(0)).
		Return(nil).
		Once()
	_, err := getDataConn(s.client, s.ftpFile, types.OpenWrite)
	s.Require().NoError(err, "no error expected")

	// brief sleep to ensure goroutines running StorFrom can all complete
	time.Sleep(50 * time.Millisecond)
}

func (s *dataConnSuite) TestGetDataConn_WriteLocationNotExistsFails() {
	someerr := errors.New("some error")
	// dataconn is nil - open for write - location doesn't exist - success
	s.client.EXPECT().
		List("/").
		Return(nil, errors.New("550")).
		Once()
	s.client.EXPECT().
		MakeDir(s.ftpFile.Location().Path()).
		Return(someerr).
		Once()
	_, err := getDataConn(s.client, s.ftpFile, types.OpenWrite)
	s.Require().ErrorIs(err, someerr, "error expected")

	// brief sleep to ensure goroutines running StorFrom can all complete
	time.Sleep(50 * time.Millisecond)
}

func (s *dataConnSuite) TestGetDataConn_ErrorWriting() {
	entries := []*ftp.Entry{{
		Name: "some",
		Type: ftp.EntryTypeFolder,
	}}
	someErr := errors.New("some error")

	// dataconn is nil - open for write - error calling client.StorFrom
	s.client.EXPECT().
		List("/").
		Return(entries, nil).
		Once()
	s.client.EXPECT().
		StorFrom(s.ftpFile.Path(), mock.Anything, uint64(0)).
		Return(someErr).
		Once()
	dc, err := getDataConn(s.client, s.ftpFile, types.OpenWrite)
	s.Require().NoError(err, "no error expected")
	// error in getDataConn should close the PipeReader meaning Write errors
	_, err = dc.Write([]byte{})
	s.Require().Error(err, "error is expected")
}

func (s *dataConnSuite) TestGetDataConn_WriteSuccess() {
	entries := []*ftp.Entry{{
		Name: "some",
		Type: ftp.EntryTypeFolder,
	}}

	// dataconn is nil - open for write - success
	s.client.EXPECT().
		List("/").
		Return(entries, nil).
		Once()
	s.client.EXPECT().
		StorFrom(s.ftpFile.Path(), mock.Anything, uint64(0)).
		Return(nil).
		Once()
	dc, err := getDataConn(s.client, s.ftpFile, types.OpenWrite)
	s.Require().NoError(err, "no error expected")
	s.IsTypef(&dataConn{}, dc, "dataconn returned")

	// brief sleep to ensure goroutines running StorFrom can all complete
	time.Sleep(50 * time.Millisecond)
}

func (s *dataConnSuite) TestGetDataConn_WriteAfterReadSuccess() {
	// open dataconn for write after dataconn for read exists
	entries := []*ftp.Entry{{
		Name: "some",
		Type: ftp.EntryTypeFolder,
	}}
	s.ftpFile.fileSystem.dataconn = &dataConn{
		mode: types.OpenRead,
		R:    io.NopCloser(strings.NewReader("")),
	}
	s.client.EXPECT().
		List("/").
		Return(entries, nil).
		Once()
	s.client.EXPECT().
		StorFrom(s.ftpFile.Path(), mock.Anything, uint64(0)).
		Return(nil).
		Once()
	dc, err := getDataConn(s.client, s.ftpFile, types.OpenWrite)
	s.Require().NoError(err, "no error expected")
	s.IsTypef(&dataConn{}, dc, "dataconn returned")

	// brief sleep to ensure goroutines running StorFrom can all complete
	time.Sleep(50 * time.Millisecond)
}

func (s *dataConnSuite) TestMode() {
	dc := &dataConn{
		mode: types.OpenRead,
	}
	s.Equal(types.OpenRead, dc.Mode())
}

func (s *dataConnSuite) TestRead() {
	contents := "some data"
	dc := &dataConn{
		R:    io.NopCloser(strings.NewReader(contents)),
		mode: types.OpenRead,
	}
	w := &strings.Builder{}
	written, err := io.Copy(w, dc)
	s.Require().NoError(err, "error not expected")
	s.Len(contents, int(written), "byte count should equal contents of reader")
	s.Equal(contents, w.String(), "read contents equals original contents")
}
