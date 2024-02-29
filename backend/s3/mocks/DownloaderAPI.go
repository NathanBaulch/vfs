// Code generated by mockery v2.39.1. DO NOT EDIT.

package mocks

import (
	context "context"
	io "io"

	mock "github.com/stretchr/testify/mock"

	s3 "github.com/aws/aws-sdk-go/service/s3"

	s3manager "github.com/aws/aws-sdk-go/service/s3/s3manager"
)

// DownloaderAPI is an autogenerated mock type for the DownloaderAPI type
type DownloaderAPI struct {
	mock.Mock
}

type DownloaderAPI_Expecter struct {
	mock *mock.Mock
}

func (_m *DownloaderAPI) EXPECT() *DownloaderAPI_Expecter {
	return &DownloaderAPI_Expecter{mock: &_m.Mock}
}

// Download provides a mock function with given fields: _a0, _a1, _a2
func (_m *DownloaderAPI) Download(_a0 io.WriterAt, _a1 *s3.GetObjectInput, _a2 ...func(*s3manager.Downloader)) (int64, error) {
	_va := make([]interface{}, len(_a2))
	for _i := range _a2 {
		_va[_i] = _a2[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, _a0, _a1)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	if len(ret) == 0 {
		panic("no return value specified for Download")
	}

	var r0 int64
	var r1 error
	if rf, ok := ret.Get(0).(func(io.WriterAt, *s3.GetObjectInput, ...func(*s3manager.Downloader)) (int64, error)); ok {
		return rf(_a0, _a1, _a2...)
	}
	if rf, ok := ret.Get(0).(func(io.WriterAt, *s3.GetObjectInput, ...func(*s3manager.Downloader)) int64); ok {
		r0 = rf(_a0, _a1, _a2...)
	} else {
		r0 = ret.Get(0).(int64)
	}

	if rf, ok := ret.Get(1).(func(io.WriterAt, *s3.GetObjectInput, ...func(*s3manager.Downloader)) error); ok {
		r1 = rf(_a0, _a1, _a2...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// DownloaderAPI_Download_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Download'
type DownloaderAPI_Download_Call struct {
	*mock.Call
}

// Download is a helper method to define mock.On call
//   - _a0 io.WriterAt
//   - _a1 *s3.GetObjectInput
//   - _a2 ...func(*s3manager.Downloader)
func (_e *DownloaderAPI_Expecter) Download(_a0 interface{}, _a1 interface{}, _a2 ...interface{}) *DownloaderAPI_Download_Call {
	return &DownloaderAPI_Download_Call{Call: _e.mock.On("Download",
		append([]interface{}{_a0, _a1}, _a2...)...)}
}

func (_c *DownloaderAPI_Download_Call) Run(run func(_a0 io.WriterAt, _a1 *s3.GetObjectInput, _a2 ...func(*s3manager.Downloader))) *DownloaderAPI_Download_Call {
	_c.Call.Run(func(args mock.Arguments) {
		variadicArgs := make([]func(*s3manager.Downloader), len(args)-2)
		for i, a := range args[2:] {
			if a != nil {
				variadicArgs[i] = a.(func(*s3manager.Downloader))
			}
		}
		run(args[0].(io.WriterAt), args[1].(*s3.GetObjectInput), variadicArgs...)
	})
	return _c
}

func (_c *DownloaderAPI_Download_Call) Return(_a0 int64, _a1 error) *DownloaderAPI_Download_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *DownloaderAPI_Download_Call) RunAndReturn(run func(io.WriterAt, *s3.GetObjectInput, ...func(*s3manager.Downloader)) (int64, error)) *DownloaderAPI_Download_Call {
	_c.Call.Return(run)
	return _c
}

// DownloadWithContext provides a mock function with given fields: _a0, _a1, _a2, _a3
func (_m *DownloaderAPI) DownloadWithContext(_a0 context.Context, _a1 io.WriterAt, _a2 *s3.GetObjectInput, _a3 ...func(*s3manager.Downloader)) (int64, error) {
	_va := make([]interface{}, len(_a3))
	for _i := range _a3 {
		_va[_i] = _a3[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, _a0, _a1, _a2)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	if len(ret) == 0 {
		panic("no return value specified for DownloadWithContext")
	}

	var r0 int64
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, io.WriterAt, *s3.GetObjectInput, ...func(*s3manager.Downloader)) (int64, error)); ok {
		return rf(_a0, _a1, _a2, _a3...)
	}
	if rf, ok := ret.Get(0).(func(context.Context, io.WriterAt, *s3.GetObjectInput, ...func(*s3manager.Downloader)) int64); ok {
		r0 = rf(_a0, _a1, _a2, _a3...)
	} else {
		r0 = ret.Get(0).(int64)
	}

	if rf, ok := ret.Get(1).(func(context.Context, io.WriterAt, *s3.GetObjectInput, ...func(*s3manager.Downloader)) error); ok {
		r1 = rf(_a0, _a1, _a2, _a3...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// DownloaderAPI_DownloadWithContext_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'DownloadWithContext'
type DownloaderAPI_DownloadWithContext_Call struct {
	*mock.Call
}

// DownloadWithContext is a helper method to define mock.On call
//   - _a0 context.Context
//   - _a1 io.WriterAt
//   - _a2 *s3.GetObjectInput
//   - _a3 ...func(*s3manager.Downloader)
func (_e *DownloaderAPI_Expecter) DownloadWithContext(_a0 interface{}, _a1 interface{}, _a2 interface{}, _a3 ...interface{}) *DownloaderAPI_DownloadWithContext_Call {
	return &DownloaderAPI_DownloadWithContext_Call{Call: _e.mock.On("DownloadWithContext",
		append([]interface{}{_a0, _a1, _a2}, _a3...)...)}
}

func (_c *DownloaderAPI_DownloadWithContext_Call) Run(run func(_a0 context.Context, _a1 io.WriterAt, _a2 *s3.GetObjectInput, _a3 ...func(*s3manager.Downloader))) *DownloaderAPI_DownloadWithContext_Call {
	_c.Call.Run(func(args mock.Arguments) {
		variadicArgs := make([]func(*s3manager.Downloader), len(args)-3)
		for i, a := range args[3:] {
			if a != nil {
				variadicArgs[i] = a.(func(*s3manager.Downloader))
			}
		}
		run(args[0].(context.Context), args[1].(io.WriterAt), args[2].(*s3.GetObjectInput), variadicArgs...)
	})
	return _c
}

func (_c *DownloaderAPI_DownloadWithContext_Call) Return(_a0 int64, _a1 error) *DownloaderAPI_DownloadWithContext_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *DownloaderAPI_DownloadWithContext_Call) RunAndReturn(run func(context.Context, io.WriterAt, *s3.GetObjectInput, ...func(*s3manager.Downloader)) (int64, error)) *DownloaderAPI_DownloadWithContext_Call {
	_c.Call.Return(run)
	return _c
}

// NewDownloaderAPI creates a new instance of DownloaderAPI. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewDownloaderAPI(t interface {
	mock.TestingT
	Cleanup(func())
}) *DownloaderAPI {
	mock := &DownloaderAPI{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}