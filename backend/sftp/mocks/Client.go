// Code generated by mockery v1.0.0. DO NOT EDIT.

package mocks

import mock "github.com/stretchr/testify/mock"
import os "os"
import pkgsftp "github.com/pkg/sftp"

import time "time"

// Client is an autogenerated mock type for the Client type
type Client struct {
	mock.Mock
}

// Chtimes provides a mock function with given fields: path, atime, mtime
func (_m *Client) Chtimes(path string, atime time.Time, mtime time.Time) error {
	ret := _m.Called(path, atime, mtime)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, time.Time, time.Time) error); ok {
		r0 = rf(path, atime, mtime)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Create provides a mock function with given fields: path
func (_m *Client) Create(path string) (*pkgsftp.File, error) {
	ret := _m.Called(path)

	var r0 *pkgsftp.File
	if rf, ok := ret.Get(0).(func(string) *pkgsftp.File); ok {
		r0 = rf(path)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*pkgsftp.File)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(path)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MkdirAll provides a mock function with given fields: path
func (_m *Client) MkdirAll(path string) error {
	ret := _m.Called(path)

	var r0 error
	if rf, ok := ret.Get(0).(func(string) error); ok {
		r0 = rf(path)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// OpenFile provides a mock function with given fields: path, f
func (_m *Client) OpenFile(path string, f int) (*pkgsftp.File, error) {
	ret := _m.Called(path, f)

	var r0 *pkgsftp.File
	if rf, ok := ret.Get(0).(func(string, int) *pkgsftp.File); ok {
		r0 = rf(path, f)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*pkgsftp.File)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, int) error); ok {
		r1 = rf(path, f)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ReadDir provides a mock function with given fields: p
func (_m *Client) ReadDir(p string) ([]os.FileInfo, error) {
	ret := _m.Called(p)

	var r0 []os.FileInfo
	if rf, ok := ret.Get(0).(func(string) []os.FileInfo); ok {
		r0 = rf(p)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]os.FileInfo)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(p)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Remove provides a mock function with given fields: path
func (_m *Client) Remove(path string) error {
	ret := _m.Called(path)

	var r0 error
	if rf, ok := ret.Get(0).(func(string) error); ok {
		r0 = rf(path)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Rename provides a mock function with given fields: oldname, newname
func (_m *Client) Rename(oldname string, newname string) error {
	ret := _m.Called(oldname, newname)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, string) error); ok {
		r0 = rf(oldname, newname)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Stat provides a mock function with given fields: p
func (_m *Client) Stat(p string) (os.FileInfo, error) {
	ret := _m.Called(p)

	var r0 os.FileInfo
	if rf, ok := ret.Get(0).(func(string) os.FileInfo); ok {
		r0 = rf(p)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(os.FileInfo)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(p)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}