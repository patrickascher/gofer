// Code generated by mockery v2.4.0-beta. DO NOT EDIT.

package mocks

import (
	cache "github.com/patrickascher/gofer/cache"
	mock "github.com/stretchr/testify/mock"

	time "time"
)

// Interface is an autogenerated mock type for the Interface type
type Interface struct {
	mock.Mock
}

// All provides a mock function with given fields:
func (_m *Interface) All() ([]cache.Item, error) {
	ret := _m.Called()

	var r0 []cache.Item
	if rf, ok := ret.Get(0).(func() []cache.Item); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]cache.Item)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Delete provides a mock function with given fields: name
func (_m *Interface) Delete(name string) error {
	ret := _m.Called(name)

	var r0 error
	if rf, ok := ret.Get(0).(func(string) error); ok {
		r0 = rf(name)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// DeleteAll provides a mock function with given fields:
func (_m *Interface) DeleteAll() error {
	ret := _m.Called()

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// GC provides a mock function with given fields:
func (_m *Interface) GC() {
	_m.Called()
}

// Get provides a mock function with given fields: name
func (_m *Interface) Get(name string) (cache.Item, error) {
	ret := _m.Called(name)

	var r0 cache.Item
	if rf, ok := ret.Get(0).(func(string) cache.Item); ok {
		r0 = rf(name)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(cache.Item)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(name)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Set provides a mock function with given fields: name, value, exp
func (_m *Interface) Set(name string, value interface{}, exp time.Duration) error {
	ret := _m.Called(name, value, exp)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, interface{}, time.Duration) error); ok {
		r0 = rf(name, value, exp)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
