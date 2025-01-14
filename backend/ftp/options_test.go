package ftp

import (
	"bytes"
	"context"
	"crypto/tls"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/c2fo/vfs/v6/utils"
)

type optionsSuite struct {
	suite.Suite
}

func TestOptions(t *testing.T) {
	suite.Run(t, &optionsSuite{})
}

func (s *optionsSuite) TestFetchUsername() {
	testCases := []struct {
		description string
		authority   string
		expected    string
	}{
		{
			description: "check defaults",
			authority:   "host.com",
			expected:    "anonymous",
		},
		{
			description: "authority value expected",
			authority:   "bob@host.com",
			expected:    "bob",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.description, func() {
			auth, err := utils.NewAuthority(tc.authority)
			s.Require().NoError(err, tc.description)

			username := fetchUsername(auth)
			s.Equal(tc.expected, username, tc.description)
		})
	}
}

func (s *optionsSuite) TestFetchPassword() {
	testCases := []struct {
		description string
		options     *Options
		envVar      *string
		expected    string
	}{
		{
			description: "check defaults",
			expected:    "anonymous",
		},
		{
			description: "env var is set but with empty value",
			expected:    "",
			envVar:      utils.Ptr(""),
		},
		{
			description: "env var is set, value should override",
			expected:    "12abc3",
			envVar:      utils.Ptr("12abc3"),
		},
		{
			description: "option should override",
			expected:    "xyz123",
			envVar:      utils.Ptr("12abc3"),
			options: &Options{
				Password: "xyz123",
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.description, func() {
			if tc.envVar != nil {
				err := os.Setenv(envPassword, *tc.envVar)
				s.Require().NoError(err, tc.description)
			}

			password := fetchPassword(tc.options)
			s.Equal(tc.expected, password, tc.description)
		})
	}
}

func (s *optionsSuite) TestFetchHostPortString() {
	testCases := []struct {
		description string
		authority   string
		envVar      *string
		expected    string
	}{
		{
			description: "check defaults",
			authority:   "user@host.com",
			expected:    "host.com:21",
		},
		{
			description: "authority has port specified",
			authority:   "user@host.com:10000",
			expected:    "host.com:10000",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.description, func() {
			auth, err := utils.NewAuthority(tc.authority)
			s.Require().NoError(err, tc.description)

			if tc.envVar != nil {
				err := os.Setenv(envPassword, *tc.envVar)
				s.Require().NoError(err, tc.description)
			}

			hostPortString := fetchHostPortString(auth)
			s.Equal(tc.expected, hostPortString, tc.description)
		})
	}
}

func (s *optionsSuite) TestIsDisableEPSV() {
	trueVal := true
	falseVal := false
	testCases := []struct {
		description string
		options     *Options
		envVar      *string
		expected    bool
	}{
		{
			description: "check defaults",
			expected:    false,
		},
		{
			description: "env var is set but empty",
			envVar:      utils.Ptr(""),
			expected:    false,
		},
		{
			description: "env var is set and is a non-true value",
			envVar:      utils.Ptr("not expected"),
			expected:    false,
		},
		{
			description: "env var is set and is a `false` value",
			envVar:      utils.Ptr("false"),
			expected:    false,
		},
		{
			description: "env var is set and is '1' value",
			envVar:      utils.Ptr("1"),
			expected:    true,
		},
		{
			description: "env var is set and is 'true'",
			envVar:      utils.Ptr("true"),
			expected:    true,
		},
		{
			description: "Options is set to false'",
			options: &Options{
				DisableEPSV: &falseVal,
			},
			expected: false,
		},
		{
			description: "Options is set to true'",
			options: &Options{
				DisableEPSV: &trueVal,
			},
			expected: true,
		},
		{
			description: "env var is set true but Options is set to false'",
			envVar:      utils.Ptr("true"),
			options: &Options{
				DisableEPSV: &falseVal,
			},
			expected: false,
		},
		{
			description: "env var is set true but Options is set to false'",
			envVar:      utils.Ptr("false"),
			options: &Options{
				DisableEPSV: &trueVal,
			},
			expected: true,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.description, func() {
			if tc.envVar != nil {
				err := os.Setenv(envDisableEPSV, *tc.envVar)
				s.Require().NoError(err, tc.description)
			}

			disabled := isDisableOption(tc.options)
			s.Equal(tc.expected, disabled, tc.description)
		})
	}
}

func (s *optionsSuite) TestFetchTLSConfig() {
	cfg := &tls.Config{
		MinVersion:             tls.VersionTLS12,
		InsecureSkipVerify:     false,
		ClientSessionCache:     tls.NewLRUClientSessionCache(0),
		ServerName:             "host.com",
		SessionTicketsDisabled: true,
	}

	testCases := []struct {
		description                string
		authority                  string
		options                    *Options
		expected                   *tls.Config
		expectInsecureCipherSuites bool
	}{
		{
			description: "check defaults",
			authority:   "user@host.com",
			expected: &tls.Config{
				MinVersion:         tls.VersionTLS12,
				InsecureSkipVerify: true, //nolint:gosec
				ClientSessionCache: tls.NewLRUClientSessionCache(0),
				ServerName:         "host.com",
			},
		},
		{
			description: "authority has port specified",
			authority:   "user@host.com:10000",
			options: &Options{
				Password:  "xyz",
				TLSConfig: cfg,
			},
			expected: cfg,
		},
		{
			description: "include insecure cipher suites",
			authority:   "user@host.com",
			options: &Options{
				IncludeInsecureCiphers: true,
			},
			expected: &tls.Config{
				MinVersion:         tls.VersionTLS12,
				InsecureSkipVerify: true, //nolint:gosec
				ClientSessionCache: tls.NewLRUClientSessionCache(0),
				ServerName:         "host.com",
			},
			expectInsecureCipherSuites: true,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.description, func() {
			auth, err := utils.NewAuthority(tc.authority)
			s.Require().NoError(err, tc.description)

			tlsCfg := fetchTLSConfig(auth, tc.options)
			s.Equal(tc.expected.MinVersion, tlsCfg.MinVersion, tc.description)
			s.Equal(tc.expected.InsecureSkipVerify, tlsCfg.InsecureSkipVerify, tc.description)
			s.Equal(tc.expected.ClientSessionCache, tlsCfg.ClientSessionCache, tc.description)
			s.Equal(tc.expected.ServerName, tlsCfg.ServerName, tc.description)
			s.Equal(tc.expected.SessionTicketsDisabled, tlsCfg.SessionTicketsDisabled, tc.description)

			if tc.expectInsecureCipherSuites {
				s.NotEmpty(tlsCfg.CipherSuites, tc.description)
				s.True(containsInsecureCipherSuites(tlsCfg.CipherSuites), tc.description)
			}
		})
	}
}

func containsInsecureCipherSuites(suites []uint16) bool {
	insecureSuites := tls.InsecureCipherSuites()
	for _, s := range suites {
		for _, insecureSuite := range insecureSuites {
			if s == insecureSuite.ID {
				return true
			}
		}
	}
	return false
}

func (s *optionsSuite) TestFetchProtocol() {
	testCases := []struct {
		description string
		options     *Options
		envVar      *string
		expected    string
	}{
		{
			description: "check defaults",
			expected:    ProtocolFTP,
		},
		{
			description: "env var is set but empty",
			envVar:      utils.Ptr(""),
			expected:    "",
		},
		{
			description: "env var is set to ftps",
			envVar:      utils.Ptr("FTPS"),
			expected:    ProtocolFTPS,
		},
		{
			description: "env var is set to ftpes",
			envVar:      utils.Ptr("FTPES"),
			expected:    ProtocolFTPES,
		},
		{
			description: "env var is set to garbage",
			envVar:      utils.Ptr("blah"),
			expected:    "blah",
		},
		{
			description: "options set to garbage",
			options: &Options{
				Protocol: ProtocolFTPS,
			},
			expected: ProtocolFTPS,
		},
		{
			description: "options set to FTPES - overriding FTPS",
			envVar:      utils.Ptr("FTPS"),
			options: &Options{
				Protocol: ProtocolFTPES,
			},
			expected: ProtocolFTPES,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.description, func() {
			s.Require().NoError(os.Unsetenv(envProtocol))
			if tc.envVar != nil {
				err := os.Setenv(envProtocol, *tc.envVar)
				s.Require().NoError(err, tc.description)
			}

			protocol := fetchProtocol(tc.options)
			s.Equal(tc.expected, protocol, tc.description)
		})
	}
}

func (s *optionsSuite) TestFetchDialOptions() {
	testCases := []struct {
		description string
		authority   string
		options     *Options
		envVar      *string
		expected    int
	}{
		{
			description: "check defaults",
			authority:   "user@host.com",
			expected:    2,
		},
		{
			description: "protocol env var is set to FTPS",
			authority:   "user@host.com",
			envVar:      utils.Ptr(ProtocolFTPS),
			expected:    3,
		},
		{
			description: "protocol env var is set to FTPES",
			authority:   "user@host.com",
			envVar:      utils.Ptr(ProtocolFTPES),
			expected:    3,
		},
		{
			description: "protocol is set to empty",
			authority:   "user@host.com",
			envVar:      utils.Ptr(""),
			expected:    2,
		},
		{
			description: "protocol Options is set to FTPS",
			authority:   "user@host.com",
			options: &Options{
				Protocol: ProtocolFTPS,
			},
			expected: 3,
		},
		{
			description: "protocol Options is set to garbage value",
			authority:   "user@host.com",
			options: &Options{
				Protocol: "blah",
			},
			expected: 2,
		},
		{
			description: "debug writer is set",
			authority:   "user@host.com",
			options: &Options{
				DebugWriter: bytes.NewBuffer([]byte{}),
			},
			expected: 3,
		},
		{
			description: "dial timeout is set",
			authority:   "user@host.com",
			options: &Options{
				DialTimeout: 1 * time.Minute,
			},
			expected: 3,
		},
		{
			description: "all options set ",
			authority:   "user@host.com",
			options: &Options{
				DebugWriter: bytes.NewBuffer([]byte{}),
				DialTimeout: 1 * time.Minute,
				Protocol:    ProtocolFTPS,
			},
			expected: 5,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.description, func() {
			if tc.envVar != nil {
				err := os.Setenv(envProtocol, *tc.envVar)
				s.Require().NoError(err, tc.description)
			}

			auth, err := utils.NewAuthority(tc.authority)
			s.Require().NoError(err, tc.description)

			dialOpts := fetchDialOptions(context.Background(), auth, tc.options)
			s.Len(dialOpts, tc.expected, tc.description)
		})
	}
}
