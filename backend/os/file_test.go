package os

import (
	"errors"
	"io"
	"os"
	"path"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/c2fo/vfs/v6"
	"github.com/c2fo/vfs/v6/mocks"
	"github.com/c2fo/vfs/v6/utils"
)

/**********************************
 ************TESTS*****************
 **********************************/

type osFileTest struct {
	suite.Suite
	testFile   vfs.File
	fileSystem *FileSystem
	tmploc     vfs.Location
}

func (s *osFileTest) SetupSuite() {
	fs := &FileSystem{}
	s.fileSystem = fs
	dir, err := os.MkdirTemp("", "os_file_test")
	s.Require().NoError(err)
	dir = utils.EnsureTrailingSlash(dir)
	s.tmploc, err = fs.NewLocation("", dir)
	s.Require().NoError(err)
	s.Require().NoError(setupTestFiles(s.tmploc))
}

func (s *osFileTest) TearDownSuite() {
	s.Require().NoError(teardownTestFiles(s.tmploc))
}

func (s *osFileTest) SetupTest() {
	file, err := s.tmploc.NewFile("test_files/test.txt")
	s.Require().NoError(err, "No file was opened")
	s.testFile = file
}

func (s *osFileTest) TearDownTest() {
	err := s.testFile.Close()
	s.Require().NoError(err, "close error not expected")
}

func (s *osFileTest) TestExists() {
	doesExist, err := s.testFile.Exists()
	s.Require().NoError(err, "Failed to check for file existence")
	s.True(doesExist)

	otherFile, err := s.tmploc.NewFile("test_files/foo.txt")
	s.Require().NoError(err, "Failed to check for file existence")

	otherFileExists, err := otherFile.Exists()
	s.Require().NoError(err)
	s.False(otherFileExists)
}

func (s *osFileTest) TestTouch() {
	// set up testfile
	testfile, err := s.tmploc.NewFile("test_files/foo.txt")
	s.Require().NoError(err)

	// testfile should NOT exist
	exists, err := testfile.Exists()
	s.Require().NoError(err)
	s.False(exists)

	// touch file
	err = testfile.Touch()
	s.Require().NoError(err)

	// testfile SHOULD exist
	exists, err = testfile.Exists()
	s.Require().NoError(err)
	s.True(exists)

	// size should be zero
	size, err := testfile.Size()
	s.Require().NoError(err)
	s.Zero(size, "size should be zero")

	// capture last_modified
	firstModTime, err := testfile.LastModified()
	s.Require().NoError(err)

	time.Sleep(time.Millisecond)

	// touch again
	err = testfile.Touch()
	s.Require().NoError(err)

	// size should still be zero
	size, err = testfile.Size()
	s.Require().NoError(err)
	s.Zero(size, "size should be zero")

	// LastModified should be later than previous LastModified
	nextModTime, err := testfile.LastModified()
	s.Require().NoError(err)
	s.True(firstModTime.Before(*nextModTime), "Last Modified was updated")
	s.Require().NoError(testfile.Close())
}

func (s *osFileTest) TestOpenFile() {
	expectedText := "hello world"
	data := make([]byte, len(expectedText))
	_, err := s.testFile.Read(data)
	s.Require().NoError(err, "read error not expected")

	s.Equal(expectedText, string(data))
}

func (s *osFileTest) TestRead() {
	// fail on nonexistent file
	noFile, err := s.tmploc.NewFile("test_files/nonexistent.txt")
	s.Require().NoError(err)
	var data []byte
	_, err = noFile.Read(data)
	s.Require().Error(err, "error trying to read nonexistent file")

	// setup file for reading
	f, err := s.tmploc.NewFile("test_files/readFile.txt")
	s.Require().NoError(err)
	b, err := f.Write([]byte("blah"))
	s.Require().NoError(err)
	s.Equal(4, b)
	s.Require().NoError(f.Close())

	// read from file
	data = make([]byte, 4)
	b, err = f.Read(data)
	s.Require().NoError(err)
	s.Equal(4, b)
	s.Equal("blah", string(data))
	s.Require().NoError(f.Close())

	// setup file for err out of opening
	f, err = s.tmploc.NewFile("test_files/readFileFail.txt")
	s.Require().NoError(err)
	f.(*File).useTempFile = true
	f.(*File).fileOpener = func(string) (*os.File, error) { return nil, errors.New("bad opener") }
	data = make([]byte, 4)
	b, err = f.Read(data)
	s.Require().Error(err)
	s.Zero(b)

	f.(*File).fileOpener = nil
	b, err = f.Write([]byte("blah"))
	s.Require().NoError(err)
	s.Equal(4, b)
	s.Require().NoError(f.Close())
	f.(*File).fileOpener = func(string) (*os.File, error) { return nil, errors.New("bad opener") }
	data = make([]byte, 4)
	b, err = f.Read(data)
	s.Require().Error(err)
	s.Zero(b)
}

func (s *osFileTest) TestSeek() {
	expectedText := "world"
	data := make([]byte, len(expectedText))
	_, err := s.testFile.Seek(6, io.SeekStart)
	s.Require().NoError(err, "seek error not expected")
	_, err = s.testFile.Read(data)
	s.Require().NoError(err, "read error not expected")
	s.Equal(expectedText, string(data))
	s.Require().NoError(s.testFile.Close())
}

func (s *osFileTest) TestCopyToLocation() {
	expectedText := "hello world"
	otherFs := mocks.NewFileSystem(s.T())
	otherFile := mocks.NewFile(s.T())

	// Expected behavior
	otherFile.EXPECT().Write([]byte(expectedText)).Return(len(expectedText), nil).Once()
	otherFile.EXPECT().Close().Return(nil)
	otherFs.EXPECT().NewFile("", "/some/path/test.txt").Return(otherFile, nil).Once()

	location := Location{name: "/some/path", fileSystem: otherFs}

	_, err := s.testFile.CopyToLocation(&location)
	s.Require().NoError(err)
}

func (s *osFileTest) TestCopyToFile() {
	expectedText := "hello world"
	otherFs := mocks.NewFileSystem(s.T())
	otherFile := mocks.NewFile(s.T())

	location := Location{name: "/some/path", fileSystem: otherFs}

	// Expected behavior
	otherFile.EXPECT().Write([]byte(expectedText)).Return(len(expectedText), nil).Once()
	otherFile.EXPECT().Close().Return(nil)
	otherFile.EXPECT().Name().Return("other.txt")
	otherFile.EXPECT().Location().Return(&location)

	otherFs.EXPECT().NewFile("", "/some/path/other.txt").Return(otherFile, nil).Once()

	err := s.testFile.CopyToFile(otherFile)
	s.Require().NoError(err)
}

func (s *osFileTest) TestEmptyCopyToFile() {
	expectedText := ""
	otherFs := mocks.NewFileSystem(s.T())
	otherFile := mocks.NewFile(s.T())

	location := Location{name: "/some/path", fileSystem: otherFs}

	// Expected behavior
	otherFile.EXPECT().Write([]byte(expectedText)).Return(len(expectedText), nil).Once().Once()
	otherFile.EXPECT().Close().Return(nil)
	otherFile.EXPECT().Name().Return("other.txt")
	otherFile.EXPECT().Location().Return(&location)

	otherFs.EXPECT().NewFile("", "/some/path/other.txt").Return(otherFile, nil).Once().Once()

	emptyFile, err := s.tmploc.NewFile("test_files/empty.txt")
	s.Require().NoError(err, "No file was opened")

	err = emptyFile.CopyToFile(otherFile)
	s.Require().NoError(err)
}

func (s *osFileTest) TestCopyToLocationIgnoreExtraSeparator() {
	expectedText := "hello world"
	otherFs := mocks.NewFileSystem(s.T())
	otherFile := mocks.NewFile(s.T())

	// Expected behavior
	otherFile.EXPECT().Write(mock.Anything).Return(len(expectedText), nil)
	otherFile.EXPECT().Close().Return(nil)
	otherFs.EXPECT().NewFile("", "/some/path/test.txt").Return(otherFile, nil).Once()

	// Add trailing slash
	location := Location{name: "/some/path/", fileSystem: otherFs}

	_, err := s.testFile.CopyToLocation(&location)
	s.Require().NoError(err)
}

func (s *osFileTest) TestMoveToLocation() {
	expectedText := "moved file"
	dir, err := os.MkdirTemp(filepath.Join(osLocationPath(s.tmploc), "test_files"), "example")
	s.Require().NoError(err)

	origFileName := filepath.Join(dir, "test_files", "move.txt")
	file, err := s.fileSystem.NewFile("", origFileName)
	s.Require().NoError(err)

	defer func() {
		err := os.RemoveAll(dir)
		s.Require().NoError(err, "remove all error not expected")
	}()

	_, err = file.Write([]byte(expectedText))
	s.Require().NoError(err, "write error not expected")

	err = file.Close()
	s.Require().NoError(err, "close error not expected")

	found, err := file.Exists()
	s.Require().NoError(err, "exists error not expected")
	s.True(found)

	// setup location
	location, err := s.fileSystem.NewLocation("", utils.EnsureTrailingSlash(dir))
	s.Require().NoError(err)

	// move the file to new location
	movedFile, err := file.MoveToLocation(location)
	s.Require().NoError(err)

	s.Equal(location.Path(), movedFile.Location().Path(), "ensure file location changed")

	// ensure the original file no longer exists
	origFile, err := s.fileSystem.NewFile(file.Location().Volume(), origFileName)
	s.Require().NoError(err)
	origFound, err := origFile.Exists()
	s.Require().NoError(err, "exists error not expected")
	s.False(origFound)

	// test non-scheme MoveToLocation
	mockLocation := mocks.NewLocation(s.T())
	mockfs := mocks.NewFileSystem(s.T())

	// Expected behavior
	mockfs.EXPECT().Scheme().Return("mock")
	fsMockFile := mocks.NewFile(s.T())
	fsMockFile.EXPECT().Write(mock.Anything).Return(10, nil)
	fsMockFile.EXPECT().Close().Return(nil)
	mockfs.EXPECT().NewFile(mock.Anything, mock.Anything).Return(fsMockFile, nil)
	mockLocation.EXPECT().FileSystem().Return(mockfs)
	mockLocation.EXPECT().Volume().Return("")
	mockLocation.EXPECT().Path().Return("/some/path/to/")
	mockFile := mocks.NewFile(s.T())
	mockFile.EXPECT().Location().Return(mockLocation)
	mockFile.EXPECT().Name().Return("/some/path/to/move.txt")
	mockFile.EXPECT().Location().Return(mockLocation)
	mockLocation.EXPECT().NewFile(mock.Anything).Return(mockFile, nil)

	_, err = movedFile.MoveToLocation(mockLocation)
	s.Require().NoError(err)
}

func (s *osFileTest) TestSafeOsRename() {
	dir, err := os.MkdirTemp(filepath.Join(osLocationPath(s.tmploc), "test_files"), "example")
	s.Require().NoError(err)
	defer func() {
		err := os.RemoveAll(dir)
		s.Require().NoError(err, "remove all error not expected")
	}()

	// TODO: I haven't figured out a way to test safeOsRename since setting up the scenario is
	//     very os-dependent so here I will actually only ensure non-"invalid cross-device link"
	//     errors are handled correctly and that normal renames work.

	// test that normal rename works
	testfile := path.Join(dir, "original.txt")
	file1, err := s.fileSystem.NewFile("", testfile)
	s.Require().NoError(err)
	testBytes := []byte("test me")
	_, err = file1.Write(testBytes)
	s.Require().NoError(err)
	s.Require().NoError(file1.Close())

	newFile := path.Join(dir, "new.txt")
	s.Require().NoError(safeOsRename(testfile, newFile))

	exists, err := file1.Exists()
	s.Require().NoError(err)
	s.False(exists)
	file2, err := s.fileSystem.NewFile("", newFile)
	s.Require().NoError(err)
	exists, err = file2.Exists()
	s.Require().NoError(err)
	s.True(exists)

	// test that a rename failure (non-"invalid cross-device link" error) is returned properly
	badfile := path.Join(dir, "this_should_not_exist.txt")

	err = safeOsRename(badfile, newFile)
	s.Require().Error(err)
	s.NotEqual(osCrossDeviceLinkError, err.Error())
}

func (s *osFileTest) TestOsCopy() {
	dir, err := os.MkdirTemp(filepath.Join(osLocationPath(s.tmploc), "test_files"), "example")
	s.Require().NoError(err)
	defer func() {
		err := os.RemoveAll(dir)
		s.Require().NoError(err, "remove all error not expected")
	}()

	file1, err := s.fileSystem.NewFile("", path.Join(dir, "original.txt"))
	s.Require().NoError(err)
	testBytes := []byte("test me")
	_, err = file1.Write(testBytes)
	s.Require().NoError(err)
	s.Require().NoError(file1.Close())

	file2, err := s.fileSystem.NewFile("", path.Join(dir, "move.txt"))
	s.Require().NoError(err)

	s.Require().NoError(osCopy(path.Join(file1.Location().Volume(), file1.Path()), path.Join(file2.Location().Volume(), file2.Path())))

	b, err := io.ReadAll(file2)
	s.Require().NoError(err)
	s.Equal(testBytes, b, "contents match")

	s.Require().NoError(file2.Close())
}

func (s *osFileTest) TestMoveToFile() {
	dir, err := os.MkdirTemp(filepath.Join(osLocationPath(s.tmploc), "test_files"), "example")
	s.Require().NoError(err)

	file1, err := s.fileSystem.NewFile("", path.Join(dir, "original.txt"))
	s.Require().NoError(err)

	file2, err := s.fileSystem.NewFile("", path.Join(dir, "move.txt"))
	s.Require().NoError(err)

	defer func() {
		err := os.RemoveAll(dir)
		s.Require().NoError(err, "remove all error not expected")
	}()

	text := "original file"
	_, err = file1.Write([]byte(text))
	s.Require().NoError(err, "write error not expected")
	err = file1.Close()
	s.Require().NoError(err, "close error not expected")

	found1, eErr1 := file1.Exists()
	s.True(found1)
	s.Require().NoError(eErr1, "exists error not expected")

	found2, eErr2 := file2.Exists()
	s.False(found2)
	s.Require().NoError(eErr2, "exists error not expected")

	err = file1.MoveToFile(file2)
	s.Require().NoError(err)

	f1Exists, err := file1.Exists()
	s.Require().NoError(err)
	f2Exists, err := file2.Exists()
	s.Require().NoError(err)
	s.False(f1Exists)
	s.True(f2Exists)

	data := make([]byte, len(text))
	_, err = file2.Read(data)
	s.Require().NoError(err, "read error not expected")
	err = file2.Close()
	s.Require().NoError(err, "close error not expected")
	s.Equal(text, string(data))

	// test non-scheme MoveToFile
	mockFile := mocks.NewFile(s.T())
	mockLocation := mocks.NewLocation(s.T())
	mockfs := mocks.NewFileSystem(s.T())

	// Expected behavior
	mockfs.EXPECT().Scheme().Return("mock")
	fsMockFile := mocks.NewFile(s.T())
	fsMockFile.EXPECT().Write(mock.Anything).Return(13, nil)
	fsMockFile.EXPECT().Close().Return(nil)
	mockfs.EXPECT().NewFile(mock.Anything, mock.Anything).Return(fsMockFile, nil)
	mockLocation.EXPECT().FileSystem().Return(mockfs)
	mockLocation.EXPECT().Volume().Return("")
	mockLocation.EXPECT().Path().Return("/some/path/to/")
	mockFile.EXPECT().Location().Return(mockLocation)
	mockFile.EXPECT().Name().Return("/some/path/to/file.txt")
	mockFile.EXPECT().Location().Return(mockLocation)

	s.Require().NoError(file2.MoveToFile(mockFile))
}

func (s *osFileTest) TestWrite() {
	expectedText := "new file"
	data := make([]byte, len(expectedText))
	file, err := s.tmploc.NewFile("test_files/new.txt")
	s.Require().NoError(err)

	_, err = file.Write([]byte(expectedText))
	s.Require().NoError(err, "write error not expected")

	_, err = file.Seek(0, io.SeekStart)
	s.Require().NoError(err, "seek error not expected")
	_, err = file.Read(data)
	s.Require().NoError(err, "read error not expected")
	err = file.Close()
	s.Require().NoError(err, "close error not expected")

	s.Equal(expectedText, string(data))

	found, err := file.Exists()
	s.Require().NoError(err, "exists error not expected")
	s.True(found)

	err = file.Delete()
	s.Require().NoError(err, "File was not deleted properly")

	found2, eErr2 := file.Exists()
	s.Require().NoError(eErr2, "exists error not expected")
	s.False(found2)

	// setup file for err out of opening
	f, err := s.tmploc.NewFile("test_files/writeFileFail.txt")
	s.Require().NoError(err)
	s.Require().NoError(f.Touch())
	_, err = f.Seek(0, io.SeekStart)
	s.Require().NoError(err)
	f.(*File).fileOpener = func(string) (*os.File, error) { return nil, errors.New("bad opener") }
	data = make([]byte, 4)
	_, err = f.Write(data)
	s.Require().Error(err)
	s.Require().NoError(f.Close())
}

func (s *osFileTest) TestCursor() {
	file, err := s.tmploc.NewFile("test_files/originalFile.txt")
	s.Require().NoError(err)

	expectedText := "mary had \na little lamb\n"
	write, err := file.Write([]byte(expectedText))
	s.Require().NoError(err, "write error not expected")
	s.Equal(24, write)
	s.Require().NoError(file.Close())

	_, err = file.Seek(5, io.SeekStart) // cursor 5 - opens fd to orig file
	s.Require().NoError(err)
	s.Equal(int64(5), file.(*File).cursorPos)

	data := make([]byte, 3)
	sz, err := file.Read(data) // cursor 8 - orig file - data: "had"
	s.Require().NoError(err)
	s.Equal(int64(8), file.(*File).cursorPos)
	s.Equal("had", string(data)) // orig file contents = "had"

	negsz := int64(-sz)
	_, serr2 := file.Seek(negsz, 1) // cursor 5 - orig file
	s.Equal(int64(5), file.(*File).cursorPos)
	s.Require().NoError(serr2)

	// because seek and/or read were called before write, write is now in in-place edit mode (not truncate-write)
	sz, err = file.Write([]byte("has")) // cursor 8 - tempfile copy of orig - write on tempfile has occurred
	s.Require().NoError(err)
	s.Equal(int64(8), file.(*File).cursorPos)
	s.Equal(3, sz)

	_, err = file.Seek(5, io.SeekStart) // cursor 5 - in temp file
	s.Require().NoError(err)
	s.Equal(int64(5), file.(*File).cursorPos)

	data = make([]byte, 3)
	sz, err = file.Read(data)
	s.Require().NoError(err)
	s.Equal(int64(8), file.(*File).cursorPos)
	s.Equal("has", string(data)) // tempFile contents = "has"
	s.Equal(3, sz)

	s.Require().NoError(file.Close()) // moves tempfile containing "has" over original file

	final := make([]byte, 8)
	rd, err := file.Read(final)
	s.Require().NoError(err)
	s.Equal(8, rd)
	s.Equal("mary has", string(final))
	s.Require().NoError(file.Close())

	// if a file exists and we overwrite with a smaller # of text, then it isn't completely overwritten
	//	//somefile.txt contains "the quick brown"
	file, err = s.tmploc.NewFile("test_files/someFile.txt")
	s.Require().NoError(err)

	expectedText = "the quick brown"
	write, err = file.Write([]byte(expectedText))
	s.Require().NoError(err, "write error not expected")
	s.Equal(15, write)

	s.Require().NoError(file.Close())

	overwrite, err := file.Write([]byte("hello")) // cursor 5 of tempfile
	s.Require().NoError(err)
	s.Equal(int64(5), file.(*File).cursorPos)
	s.Equal(5, overwrite)

	data = make([]byte, 5)
	_, err = file.Seek(0, io.SeekStart) // cursor 0 of tempfile
	s.Require().NoError(err)

	_, err = file.Read(data) // cursor 5 of tempfile - data: "hello"
	s.Require().NoError(err)
	s.Equal("hello", string(data))

	data = make([]byte, 3)
	sought, err := file.Seek(-3, 2) // cursor 3 from end of tempfile
	s.Require().NoError(err)
	s.Equal(int64(2), sought) // seek returns position relative to beginning of file

	rd, err = file.Read(data) // cursor 0 from end of tempfile - data: "llo"
	s.Require().NoError(err)
	s.Equal(3, rd)
	s.Equal("llo", string(data))
	s.Equal(int64(5), file.(*File).cursorPos)

	_, err = file.Seek(0, io.SeekStart)
	s.Require().NoError(err)
	final = make([]byte, 5)
	rd, err = file.Read(final)
	s.Require().NoError(err)
	s.Equal(5, rd)
	s.Equal("hello", string(final))
	s.Require().NoError(file.Close()) // moves tempfile containing "hello" over somefile.txt
}

func (s *osFileTest) TestCursorErrs() {
	noFile, err := s.tmploc.NewFile("test_files/nonet.txt")
	s.Require().NoError(err)
	data := make([]byte, 10)
	_, err = noFile.Read(data)
	s.Require().Error(err)
	s.Require().NoError(noFile.Close())

	_, err = noFile.Seek(10, 10)
	s.Require().Error(err)

	_, err = noFile.Seek(-10, 2)
	s.Require().Error(err)
	s.Require().NoError(noFile.Close())

	noFile.(*File).fileOpener = nil
	noFile.(*File).useTempFile = false
	b, err := noFile.Write([]byte("blah"))
	s.Require().NoError(err)
	s.Equal(4, b)
	noFile.(*File).fileOpener = func(string) (*os.File, error) { return nil, errors.New("bad opener") }
	s.Require().Error(noFile.Close())
}

func (s *osFileTest) TestLastModified() {
	file, err := s.tmploc.NewFile("test_files/test.txt")
	s.Require().NoError(err)

	lastModified, err := file.LastModified()
	s.Require().NoError(err)
	osStats, err := os.Stat(filepath.Join(osLocationPath(s.tmploc), "test_files", "test.txt"))
	s.Require().NoError(err)
	s.NotNil(lastModified)
	s.Equal(osStats.ModTime(), *lastModified)

	noFile, err := s.tmploc.NewFile("test_files/nonexistent.txt")
	s.Require().NoError(err)
	lastModified, err = noFile.LastModified()
	s.Require().Error(err)
	s.Nil(lastModified)
}

func (s *osFileTest) TestName() {
	file, err := s.tmploc.NewFile("test_files/test.txt")
	s.Require().NoError(err)
	s.Equal("test.txt", file.Name())
}

func (s *osFileTest) TestSize() {
	file, err := s.tmploc.NewFile("test_files/test.txt")
	s.Require().NoError(err)

	size, err := file.Size()
	s.Require().NoError(err)

	osStats, err := os.Stat(filepath.Join(osLocationPath(s.tmploc), "test_files", "test.txt"))
	s.Require().NoError(err)
	s.NotNil(size)
	s.Equal(osStats.Size(), int64(size))

	noFile, err := s.tmploc.NewFile("test_files/nonexistent.txt")
	s.Require().NoError(err)
	size, err = noFile.Size()
	s.Require().Error(err)
	s.Zero(size)
}

func (s *osFileTest) TestPath() {
	file, err := s.tmploc.NewFile("test_files/test.txt")
	s.Require().NoError(err)
	s.Equal(path.Join(file.Location().Path(), file.Name()), file.Path())
}

func (s *osFileTest) TestURI() {
	file, err := s.tmploc.NewFile("some/file/test.txt")
	s.Require().NoError(err)
	expected := "file://" + filepath.ToSlash(filepath.Join(osLocationPath(s.tmploc), "some", "file", "test.txt"))
	s.Equal(expected, file.URI(), "%s does not match %s", file.URI(), expected)
}

func (s *osFileTest) TestStringer() {
	file, err := s.tmploc.NewFile("some/file/test.txt")
	s.Require().NoError(err)
	s.Equal("file://"+filepath.ToSlash(filepath.Join(osLocationPath(s.tmploc), "some", "file", "test.txt")), file.String())
}

func (s *osFileTest) TestLocationRightAfterChangeDir() {
	file, err := s.tmploc.NewFile("chdTest.txt")
	s.Require().NoError(err)
	chDir := "someDir/"

	loc := file.Location()
	s.NotContains(loc.Path(), "someDir/", "location should not contain 'someDir/'")

	err = loc.ChangeDir(chDir)
	s.Require().NoError(err)
	s.Contains(loc.Path(), "someDir/", "location now should contain 'someDir/'")

	// file location shouldn't be affected by ChangeDir() on Location
	s.NotContains(file.Location().Path(), "someDir/", "file location should NOT contain 'someDir/'")
}

func TestOSFile(t *testing.T) {
	suite.Run(t, &osFileTest{})
	_ = os.Remove("test_files/new.txt")
}

/*
Setup TEST FILES
*/
func setupTestFiles(baseLoc vfs.Location) error {
	// setup "test_files" dir
	if err := createDir(baseLoc, "test_files"); err != nil {
		return err
	}

	// setup "test_files/test.txt"
	if err := writeStringFile(baseLoc, "test_files/empty.txt", ``); err != nil {
		return err
	}

	// setup "test_files/test.txt"
	if err := writeStringFile(baseLoc, "test_files/prefix-file.txt", `hello, Dave`); err != nil {
		return err
	}

	// setup "test_files/test.txt"
	if err := writeStringFile(baseLoc, "test_files/test.txt", `hello world`); err != nil {
		return err
	}

	// setup "test_files/subdir" dir
	if err := createDir(baseLoc, "test_files/subdir"); err != nil {
		return err
	}

	// setup "test_files/subdir/test.txt"
	if err := writeStringFile(baseLoc, "test_files/subdir/test.txt", `hello world too`); err != nil {
		return err
	}

	return nil
}

func teardownTestFiles(baseLoc vfs.Location) error {
	return os.RemoveAll(baseLoc.Path())
}

func createDir(baseLoc vfs.Location, dirname string) error {
	dir := path.Join(baseLoc.Path(), dirname)
	perm := os.FileMode(0o755)
	err := os.Mkdir(dir, perm)
	if err != nil {
		_ = teardownTestFiles(baseLoc)
	}
	return err
}

func writeStringFile(baseLoc vfs.Location, filename, data string) error {
	file := path.Join(baseLoc.Path(), filename)
	f, err := os.Create(file) //nolint:gosec
	if err != nil {
		_ = teardownTestFiles(baseLoc)
		return err
	}
	_, err = f.WriteString(data)
	if err != nil {
		_ = teardownTestFiles(baseLoc)
		return err
	}
	err = f.Close()
	if err != nil {
		_ = teardownTestFiles(baseLoc)
		return err
	}
	return nil
}
