// Code generated by mockery v2.4.0-beta. DO NOT EDIT.

package middleware_test

import mock "github.com/stretchr/testify/mock"

// RoleService is an autogenerated mock type for the RoleService type
type RoleService struct {
	mock.Mock
}

// Allowed provides a mock function with given fields: pattern, HTTPMethod, claims
func (_m *RoleService) Allowed(pattern string, HTTPMethod string, claims interface{}) bool {
	ret := _m.Called(pattern, HTTPMethod, claims)

	var r0 bool
	if rf, ok := ret.Get(0).(func(string, string, interface{}) bool); ok {
		r0 = rf(pattern, HTTPMethod, claims)
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}
