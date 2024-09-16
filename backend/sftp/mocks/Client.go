// Code generated by mockery v2.43.2. DO NOT EDIT.

package mocks

import (
	fs "io/fs"

	pkgsftp "github.com/pkg/sftp"
	mock "github.com/stretchr/testify/mock"

	time "time"
)

// Client is an autogenerated mock type for the Client type
type Client struct {
	mock.Mock
}

type Client_Expecter struct {
	mock *mock.Mock
}

func (_m *Client) EXPECT() *Client_Expecter {
	return &Client_Expecter{mock: &_m.Mock}
}

// Chmod provides a mock function with given fields: path, mode
func (_m *Client) Chmod(path string, mode fs.FileMode) error {
	ret := _m.Called(path, mode)

	if len(ret) == 0 {
		panic("no return value specified for Chmod")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(string, fs.FileMode) error); ok {
		r0 = rf(path, mode)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Client_Chmod_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Chmod'
type Client_Chmod_Call struct {
	*mock.Call
}

// Chmod is a helper method to define mock.On call
//   - path string
//   - mode fs.FileMode
func (_e *Client_Expecter) Chmod(path interface{}, mode interface{}) *Client_Chmod_Call {
	return &Client_Chmod_Call{Call: _e.mock.On("Chmod", path, mode)}
}

func (_c *Client_Chmod_Call) Run(run func(path string, mode fs.FileMode)) *Client_Chmod_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string), args[1].(fs.FileMode))
	})
	return _c
}

func (_c *Client_Chmod_Call) Return(_a0 error) *Client_Chmod_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Client_Chmod_Call) RunAndReturn(run func(string, fs.FileMode) error) *Client_Chmod_Call {
	_c.Call.Return(run)
	return _c
}

// Chtimes provides a mock function with given fields: path, atime, mtime
func (_m *Client) Chtimes(path string, atime time.Time, mtime time.Time) error {
	ret := _m.Called(path, atime, mtime)

	if len(ret) == 0 {
		panic("no return value specified for Chtimes")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(string, time.Time, time.Time) error); ok {
		r0 = rf(path, atime, mtime)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Client_Chtimes_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Chtimes'
type Client_Chtimes_Call struct {
	*mock.Call
}

// Chtimes is a helper method to define mock.On call
//   - path string
//   - atime time.Time
//   - mtime time.Time
func (_e *Client_Expecter) Chtimes(path interface{}, atime interface{}, mtime interface{}) *Client_Chtimes_Call {
	return &Client_Chtimes_Call{Call: _e.mock.On("Chtimes", path, atime, mtime)}
}

func (_c *Client_Chtimes_Call) Run(run func(path string, atime time.Time, mtime time.Time)) *Client_Chtimes_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string), args[1].(time.Time), args[2].(time.Time))
	})
	return _c
}

func (_c *Client_Chtimes_Call) Return(_a0 error) *Client_Chtimes_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Client_Chtimes_Call) RunAndReturn(run func(string, time.Time, time.Time) error) *Client_Chtimes_Call {
	_c.Call.Return(run)
	return _c
}

// Close provides a mock function with given fields:
func (_m *Client) Close() error {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for Close")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Client_Close_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Close'
type Client_Close_Call struct {
	*mock.Call
}

// Close is a helper method to define mock.On call
func (_e *Client_Expecter) Close() *Client_Close_Call {
	return &Client_Close_Call{Call: _e.mock.On("Close")}
}

func (_c *Client_Close_Call) Run(run func()) *Client_Close_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Client_Close_Call) Return(_a0 error) *Client_Close_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Client_Close_Call) RunAndReturn(run func() error) *Client_Close_Call {
	_c.Call.Return(run)
	return _c
}

// Create provides a mock function with given fields: path
func (_m *Client) Create(path string) (*pkgsftp.File, error) {
	ret := _m.Called(path)

	if len(ret) == 0 {
		panic("no return value specified for Create")
	}

	var r0 *pkgsftp.File
	var r1 error
	if rf, ok := ret.Get(0).(func(string) (*pkgsftp.File, error)); ok {
		return rf(path)
	}
	if rf, ok := ret.Get(0).(func(string) *pkgsftp.File); ok {
		r0 = rf(path)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*pkgsftp.File)
		}
	}

	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(path)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Client_Create_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Create'
type Client_Create_Call struct {
	*mock.Call
}

// Create is a helper method to define mock.On call
//   - path string
func (_e *Client_Expecter) Create(path interface{}) *Client_Create_Call {
	return &Client_Create_Call{Call: _e.mock.On("Create", path)}
}

func (_c *Client_Create_Call) Run(run func(path string)) *Client_Create_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string))
	})
	return _c
}

func (_c *Client_Create_Call) Return(_a0 *pkgsftp.File, _a1 error) *Client_Create_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *Client_Create_Call) RunAndReturn(run func(string) (*pkgsftp.File, error)) *Client_Create_Call {
	_c.Call.Return(run)
	return _c
}

// MkdirAll provides a mock function with given fields: path
func (_m *Client) MkdirAll(path string) error {
	ret := _m.Called(path)

	if len(ret) == 0 {
		panic("no return value specified for MkdirAll")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(string) error); ok {
		r0 = rf(path)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Client_MkdirAll_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'MkdirAll'
type Client_MkdirAll_Call struct {
	*mock.Call
}

// MkdirAll is a helper method to define mock.On call
//   - path string
func (_e *Client_Expecter) MkdirAll(path interface{}) *Client_MkdirAll_Call {
	return &Client_MkdirAll_Call{Call: _e.mock.On("MkdirAll", path)}
}

func (_c *Client_MkdirAll_Call) Run(run func(path string)) *Client_MkdirAll_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string))
	})
	return _c
}

func (_c *Client_MkdirAll_Call) Return(_a0 error) *Client_MkdirAll_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Client_MkdirAll_Call) RunAndReturn(run func(string) error) *Client_MkdirAll_Call {
	_c.Call.Return(run)
	return _c
}

// OpenFile provides a mock function with given fields: path, f
func (_m *Client) OpenFile(path string, f int) (*pkgsftp.File, error) {
	ret := _m.Called(path, f)

	if len(ret) == 0 {
		panic("no return value specified for OpenFile")
	}

	var r0 *pkgsftp.File
	var r1 error
	if rf, ok := ret.Get(0).(func(string, int) (*pkgsftp.File, error)); ok {
		return rf(path, f)
	}
	if rf, ok := ret.Get(0).(func(string, int) *pkgsftp.File); ok {
		r0 = rf(path, f)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*pkgsftp.File)
		}
	}

	if rf, ok := ret.Get(1).(func(string, int) error); ok {
		r1 = rf(path, f)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Client_OpenFile_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'OpenFile'
type Client_OpenFile_Call struct {
	*mock.Call
}

// OpenFile is a helper method to define mock.On call
//   - path string
//   - f int
func (_e *Client_Expecter) OpenFile(path interface{}, f interface{}) *Client_OpenFile_Call {
	return &Client_OpenFile_Call{Call: _e.mock.On("OpenFile", path, f)}
}

func (_c *Client_OpenFile_Call) Run(run func(path string, f int)) *Client_OpenFile_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string), args[1].(int))
	})
	return _c
}

func (_c *Client_OpenFile_Call) Return(_a0 *pkgsftp.File, _a1 error) *Client_OpenFile_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *Client_OpenFile_Call) RunAndReturn(run func(string, int) (*pkgsftp.File, error)) *Client_OpenFile_Call {
	_c.Call.Return(run)
	return _c
}

// ReadDir provides a mock function with given fields: p
func (_m *Client) ReadDir(p string) ([]fs.FileInfo, error) {
	ret := _m.Called(p)

	if len(ret) == 0 {
		panic("no return value specified for ReadDir")
	}

	var r0 []fs.FileInfo
	var r1 error
	if rf, ok := ret.Get(0).(func(string) ([]fs.FileInfo, error)); ok {
		return rf(p)
	}
	if rf, ok := ret.Get(0).(func(string) []fs.FileInfo); ok {
		r0 = rf(p)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]fs.FileInfo)
		}
	}

	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(p)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Client_ReadDir_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ReadDir'
type Client_ReadDir_Call struct {
	*mock.Call
}

// ReadDir is a helper method to define mock.On call
//   - p string
func (_e *Client_Expecter) ReadDir(p interface{}) *Client_ReadDir_Call {
	return &Client_ReadDir_Call{Call: _e.mock.On("ReadDir", p)}
}

func (_c *Client_ReadDir_Call) Run(run func(p string)) *Client_ReadDir_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string))
	})
	return _c
}

func (_c *Client_ReadDir_Call) Return(_a0 []fs.FileInfo, _a1 error) *Client_ReadDir_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *Client_ReadDir_Call) RunAndReturn(run func(string) ([]fs.FileInfo, error)) *Client_ReadDir_Call {
	_c.Call.Return(run)
	return _c
}

// Remove provides a mock function with given fields: path
func (_m *Client) Remove(path string) error {
	ret := _m.Called(path)

	if len(ret) == 0 {
		panic("no return value specified for Remove")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(string) error); ok {
		r0 = rf(path)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Client_Remove_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Remove'
type Client_Remove_Call struct {
	*mock.Call
}

// Remove is a helper method to define mock.On call
//   - path string
func (_e *Client_Expecter) Remove(path interface{}) *Client_Remove_Call {
	return &Client_Remove_Call{Call: _e.mock.On("Remove", path)}
}

func (_c *Client_Remove_Call) Run(run func(path string)) *Client_Remove_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string))
	})
	return _c
}

func (_c *Client_Remove_Call) Return(_a0 error) *Client_Remove_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Client_Remove_Call) RunAndReturn(run func(string) error) *Client_Remove_Call {
	_c.Call.Return(run)
	return _c
}

// Rename provides a mock function with given fields: oldname, newname
func (_m *Client) Rename(oldname string, newname string) error {
	ret := _m.Called(oldname, newname)

	if len(ret) == 0 {
		panic("no return value specified for Rename")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(string, string) error); ok {
		r0 = rf(oldname, newname)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Client_Rename_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Rename'
type Client_Rename_Call struct {
	*mock.Call
}

// Rename is a helper method to define mock.On call
//   - oldname string
//   - newname string
func (_e *Client_Expecter) Rename(oldname interface{}, newname interface{}) *Client_Rename_Call {
	return &Client_Rename_Call{Call: _e.mock.On("Rename", oldname, newname)}
}

func (_c *Client_Rename_Call) Run(run func(oldname string, newname string)) *Client_Rename_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string), args[1].(string))
	})
	return _c
}

func (_c *Client_Rename_Call) Return(_a0 error) *Client_Rename_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Client_Rename_Call) RunAndReturn(run func(string, string) error) *Client_Rename_Call {
	_c.Call.Return(run)
	return _c
}

// Stat provides a mock function with given fields: p
func (_m *Client) Stat(p string) (fs.FileInfo, error) {
	ret := _m.Called(p)

	if len(ret) == 0 {
		panic("no return value specified for Stat")
	}

	var r0 fs.FileInfo
	var r1 error
	if rf, ok := ret.Get(0).(func(string) (fs.FileInfo, error)); ok {
		return rf(p)
	}
	if rf, ok := ret.Get(0).(func(string) fs.FileInfo); ok {
		r0 = rf(p)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(fs.FileInfo)
		}
	}

	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(p)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Client_Stat_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Stat'
type Client_Stat_Call struct {
	*mock.Call
}

// Stat is a helper method to define mock.On call
//   - p string
func (_e *Client_Expecter) Stat(p interface{}) *Client_Stat_Call {
	return &Client_Stat_Call{Call: _e.mock.On("Stat", p)}
}

func (_c *Client_Stat_Call) Run(run func(p string)) *Client_Stat_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string))
	})
	return _c
}

func (_c *Client_Stat_Call) Return(_a0 fs.FileInfo, _a1 error) *Client_Stat_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *Client_Stat_Call) RunAndReturn(run func(string) (fs.FileInfo, error)) *Client_Stat_Call {
	_c.Call.Return(run)
	return _c
}

// NewClient creates a new instance of Client. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewClient(t interface {
	mock.TestingT
	Cleanup(func())
}) *Client {
	mock := &Client{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
