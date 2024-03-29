// Code generated by mockery v2.33.1. DO NOT EDIT.

package automock

import mock "github.com/stretchr/testify/mock"

// BucketObject is an autogenerated mock type for the BucketObject type
type BucketObject struct {
	mock.Mock
}

// Bucket provides a mock function with given fields:
func (_m *BucketObject) Bucket() string {
	ret := _m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// Name provides a mock function with given fields:
func (_m *BucketObject) Name() string {
	ret := _m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// NewBucketObject creates a new instance of BucketObject. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewBucketObject(t interface {
	mock.TestingT
	Cleanup(func())
}) *BucketObject {
	mock := &BucketObject{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
