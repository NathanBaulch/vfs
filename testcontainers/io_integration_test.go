package testcontainers

import (
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/c2fo/vfs/v6"
	"github.com/c2fo/vfs/v6/vfssimple"
)

type osWrapper struct {
	filename   string
	file       *os.File
	exists     bool
	seekCalled bool
}

type ioTestSuite struct {
	suite.Suite
	testLocations map[string]vfs.Location
}

func (s *ioTestSuite) SetupSuite() {
	registers := []func(*testing.T) string{
		//registerMem,
		//registerOS,
		//registerAtmoz,
		//registerAzurite,
		//registerGCSServer,
		////registerLocalStack,
		////registerMinio,
		registerVSFTPC,
	}
	uris := make([]string, len(registers))
	var wg sync.WaitGroup
	for i := range registers {
		wg.Add(1)
		go func() {
			uris[i] = registers[i](s.T())
			wg.Done()
		}()
	}
	wg.Wait()

	s.testLocations = make(map[string]vfs.Location)
	for _, u := range uris {
		l, err := vfssimple.NewLocation(u)
		s.Require().NoError(err)
		s.testLocations[l.FileSystem().Scheme()] = l
	}
}

func (s *ioTestSuite) TestFileOperations() {
	for scheme, location := range s.testLocations {
		s.Run(scheme, func() {
			s.testFileOperations(location.URI())
		})
	}
}

// unless seek or read is called first, writes should replace a file (not edit)

func (s *ioTestSuite) testFileOperations(testPath string) {
	testCases := []struct {
		description       string
		sequence          string
		fileAlreadyExists bool
		expectFailure     bool
		expectedResults   string
	}{
		// Read, Close file
		{
			"Read, Close, file exists",
			"R(all);C()",
			true,
			false,
			"some text",
		},
		{
			"Read, Close, file does not exist",
			"R(all);C()",
			false,
			true,
			"",
		},

		// Read, Seek, Read, Close
		{
			"Read, Seek, Read, Close, file exists",
			"R(4);S(0,0);R(4);C()",
			true,
			false,
			"some text",
		},

		// Write, Close
		{
			"Write, Close, file does not exist",
			"W(abc);C()",
			false,
			false,
			"abc",
		},
		{
			"Write, Close, file exists",
			"W(abc);C()",
			true,
			false,
			"abc",
		},

		// Write, Seek, Write, Close
		{
			"Write, Seek, Write, Close, file does not exist",
			"W(this and that);S(0,0);W(that);C()",
			false,
			false,
			"that and that",
		},
		{
			"Write, Seek, Write, Close, file exists",
			"W(this and that);S(0,0);W(that);C()",
			true,
			false,
			"that and that",
		},

		// Seek
		{
			"Seek, Close - file does not exist",
			"S(2,0);C()",
			false,
			true,
			"",
		},
		{
			"Seek, Close - file exists",
			"S(2,0);C()",
			true,
			false,
			"some text",
		},
		{
			"Seek, Write, Close, file exists",
			"S(5,0);W(new text);C()",
			true,
			false,
			"some new text",
		},

		// Seek, Read, Close
		{
			"Seek, Read, Close, file does not exist",
			"S(5,0);R(4);C()",
			false,
			true,
			"",
		},
		{
			"Seek, Read, Close, file exists",
			"S(5,0);R(4);C()",
			true,
			false,
			"some text",
		},

		// Read, Write, Close
		{
			"Read, Write, Close, file does not exist",
			"R(5);W(new text);C()",
			false,
			true,
			"",
		},
		{
			"Read, Write, Close, file exists",
			"R(5);W(new text);C()",
			true,
			false,
			"some new text",
		},

		// Read, Seek, Write, Close
		{
			"Read, Seek, Write, Close, file does not exist",
			"R(2);S(3,1);W(new text);C()",
			false,
			true,
			"",
		},
		{
			"Read, Seek, Write, Close, file exists",
			"R(2);S(3,1);W(new text);C()",
			true,
			false,
			"some new text",
		},

		// Write, Seek, Read, Close
		{
			"Write, Seek, Read, Close, file does not exist",
			"W(new text);S(0,0);R(5);C()",
			false,
			false,
			"new text",
		},
		{
			"Write, Seek, Read, Close, file exists",
			"W(new text);S(0,0);R(5);C()",
			true,
			false,
			"new text",
		},
	}

	defer s.teardownTestLocation(testPath)
	for _, tc := range testCases {
		if tc.description != "Write, Seek, Write, Close, file exists" {
			continue
		}
		s.Run(tc.description, func() {
			testFileName := "testfile.txt"

			// Setup vfs environment
			file, err := s.setupTestFile(tc.fileAlreadyExists, testPath, testFileName) // Implement this setup function
			defer func() {
				if file != nil {
					_ = file.Close()
					_ = file.Delete()
				}
			}()
			s.Require().NoError(err)

			// Use vfs to execute the sequence of operations described by the description
			actualContents, err := s.executeSequence(file, tc.sequence) // Implement this function

			// Assert expected outcomes
			if tc.expectFailure {
				s.Error(err, "%s: expected failure but got success", tc.description)
			} else {
				s.Require().NoError(err, "%s: expected success but got failure", tc.description)
			}

			s.Equal(tc.expectedResults, actualContents, "%s: expected results %s but got %s", tc.description, tc.expectedResults, actualContents)
		})
	}
}

//nolint:gocyclo
func (s *ioTestSuite) executeSequence(file vfs.File, sequence string) (string, error) {
	// split sequence by semicolon
	commands := strings.Split(sequence, ";")
	var commandErr error
SEQ:
	for _, command := range commands {
		// parse command
		commandName, commandArgs := s.parseCommand(command)

		switch commandName {
		case "R":
			if commandArgs[0] == "all" {
				// Read entire file
				_, commandErr = io.ReadAll(file)
				if commandErr != nil {
					break SEQ
				}
			} else {
				// convert arg 0 to uint64
				bytesize, err := strconv.ParseUint(commandArgs[0], 10, 64)
				s.Require().NoError(err, "invalid bytesize: %s", commandArgs[0])

				// Read file
				b := make([]byte, bytesize)
				_, commandErr = file.Read(b)
				if commandErr != nil {
					break SEQ
				}
			}
		case "W":
			// Write to file
			_, commandErr = file.Write([]byte(commandArgs[0]))
			if commandErr != nil {
				break SEQ
			}
		case "S":
			// expect 2 args for offset and whence
			s.Require().Len(commandArgs, 2, "invalid number of args for Seek: %d", len(commandArgs))
			// convert args
			offset, err := strconv.ParseInt(commandArgs[0], 10, 64)
			s.Require().NoError(err, "invalid offset: %s", commandArgs[0])
			whence, err := strconv.Atoi(commandArgs[1])
			s.Require().NoError(err, "invalid whence: %s", commandArgs[1])
			// Seek
			_, commandErr = file.Seek(offset, whence)
			if commandErr != nil {
				break SEQ
			}
		case "C":
			// Close
			commandErr = file.Close()
			if commandErr != nil {
				break SEQ
			}
		}
	}
	// success so compare file contents to expected results
	if commandErr != nil {
		return "", commandErr
	}

	var f io.ReadCloser

	var err error
	f, err = file.Location().NewFile(file.Name())
	s.Require().NoError(err, "error opening file")
	defer func() { _ = f.Close() }()
	// Read entire file
	contents, err := io.ReadAll(f)
	s.Require().NoError(err, "error reading file")
	return string(contents), nil
}

var commandArgsRegex = regexp.MustCompile(`^([a-zA-Z0-9]+)\((.*)\)$`)

// takes command string in the form of <command name>(<args>) and returns the command name and args
func (s *ioTestSuite) parseCommand(command string) (string, []string) {
	// parse command string
	results := commandArgsRegex.FindStringSubmatch(command)
	s.Require().Len(results, 3, "invalid command string: %s", command)

	// split args by comma
	args := strings.Split(results[2], ",")

	return results[1], args
}

func (s *ioTestSuite) setupTestFile(existsBefore bool, loc, filename string) (vfs.File, error) {
	// Create file
	scheme, _, _ := strings.Cut(loc, ":")
	// Write something to the file
	f, err := s.testLocations[scheme].NewFile(filename)
	if err != nil {
		return nil, err
	}
	if existsBefore {
		_, err = f.Write([]byte("some text"))
		if err != nil {
			return nil, err
		}
		err = f.Close()
		if err != nil {
			return nil, err
		}
	}

	return f, nil
}

func (s *ioTestSuite) teardownTestLocation(testPath string) {
	scheme, _, _ := strings.Cut(testPath, ":")
	// Write something to the file
	loc := s.testLocations[scheme]
	files, err := loc.List()
	s.Require().NoError(err)
	for _, file := range files {
		err := loc.DeleteFile(file)
		s.Require().NoError(err)
	}
}

func TestIOTestSuite(t *testing.T) {
	suite.Run(t, &ioTestSuite{})
}
