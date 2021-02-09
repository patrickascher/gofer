// Copyright (c) 2021 Patrick Ascher <development@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package cache

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// TestManager runs all tests.
func TestManager(t *testing.T) {
	asserts := assert.New(t)
	mockProvider := new(MockInterface)
	mockProvider.On("GC").Once()

	// register the cache provider.
	err := Register("all", func(o interface{}) (Interface, error) { return mockProvider, nil })
	asserts.NoError(err)

	// get the cache manager over the registry.
	m, err := New("all", nil)
	asserts.NoError(err)

	managerStruct := m.(*manager)

	// setting some manager default values
	managerStruct.SetDefaultPrefix("testing")
	managerStruct.SetDefaultExpiration(3 * time.Hour)

	// test cases
	testingManagerSet(asserts, managerStruct, mockProvider)
	testingManagerAll(asserts, managerStruct, mockProvider)
	testingManagerGet(asserts, managerStruct, mockProvider)
	// testing hit & miss counter
	asserts.Equal(4, managerStruct.HitCount(DefaultPrefix, "foo"))
	asserts.Equal(5, managerStruct.HitCount("names", "john"))
	asserts.Equal(1, managerStruct.HitCount("names", "john3"))
	asserts.Equal(1, managerStruct.MissCount("names", "john3"))
	testingManagerExist(asserts, managerStruct, mockProvider)
	testingManagerPrefix(asserts, managerStruct, mockProvider)
	testingManagerDelete(asserts, managerStruct, mockProvider)
	testingManagerDeletePrefix(asserts, managerStruct, mockProvider)
	testingManagerDeleteAll(asserts, managerStruct, mockProvider)

	// needed because GC() is a goroutine
	time.Sleep(10 * time.Millisecond)
	// check the mock expectations
	mockProvider.AssertExpectations(t)
}

// testingManagerSet is testing:
// - if cache.DefaultPrefix are set correctly
// - if cache.NoExpiration is set correctly
// - if cache.DefaultExpiration is manipulated correctly.
// - error handling if provider returns one.
// - if the prefix struct is created correctly.
// - if the statistic struct is created correctly.
// - if re-assigning a cache item is working like planned.
func testingManagerSet(asserts *assert.Assertions, managerStruct *manager, mockProvider *MockInterface) {
	// testing set with the default prefix and no expiration time
	mockProvider.On("Set", managerStruct.prefixedName(DefaultPrefix, "foo"), "bar", time.Duration(NoExpiration)).Once().Return(nil)
	err := managerStruct.Set(DefaultPrefix, "foo", "bar", NoExpiration)
	asserts.NoError(err)

	// error: set provider returns one.
	mockProvider.On("Set", managerStruct.prefixedName(DefaultPrefix, "foo"), "bar", time.Duration(NoExpiration)).Once().Return(errors.New("an error"))
	err = managerStruct.Set(DefaultPrefix, "foo", "bar", NoExpiration)
	asserts.Error(err)
	asserts.Equal("an error", errors.Unwrap(err).Error())

	// testing set with the default prefix and no expiration time
	mockProvider.On("Set", managerStruct.prefixedName("names", "john"), "doe", managerStruct.defaultExpiration).Once().Return(nil)
	err = managerStruct.Set("names", "john", "doe", DefaultExpiration)
	asserts.NoError(err)

	// testing the prefix struct
	expectedPrefixMap := map[string][]string{"": {"foo"}, "names": {"john"}}
	asserts.True(fmt.Sprint(expectedPrefixMap) == fmt.Sprint(managerStruct.prefixes))
	// check if all (foo, john) statistics are created
	expectedStatisticsMap := map[string]counter{"testing_foo": {hit: 0, miss: 0, exists: true}, "names_john": {hit: 0, miss: 0, exists: true}}
	asserts.True(fmt.Sprint(expectedStatisticsMap) == fmt.Sprint(managerStruct.statistics))
	// re-assign an already existing kex
	mockProvider.On("Set", managerStruct.prefixedName("names", "john"), "doe", managerStruct.defaultExpiration).Once().Return(nil)
	err = managerStruct.Set("names", "john", "doe", DefaultExpiration)
	asserts.NoError(err)
	// adding new key
	mockProvider.On("Set", managerStruct.prefixedName("names", "john2"), "doe", managerStruct.defaultExpiration).Once().Return(nil)
	err = managerStruct.Set("names", "john2", "doe", DefaultExpiration)
	asserts.NoError(err)

	// testing the prefix struct
	expectedPrefixMap = map[string][]string{"": {"foo"}, "names": {"john", "john2"}}
	asserts.True(fmt.Sprint(expectedPrefixMap) == fmt.Sprint(managerStruct.prefixes))
	// check if all (foo, john, john2) statistics are created
	expectedStatisticsMap = map[string]counter{"testing_foo": {hit: 0, miss: 0, exists: true}, "names_john": {hit: 0, miss: 0, exists: true}, "names_john2": {hit: 0, miss: 0, exists: true}}
	asserts.True(fmt.Sprint(expectedStatisticsMap) == fmt.Sprint(managerStruct.statistics))
}

// testingManagerAll tests
// - if All() returns the provider return.
// - if the statistic will increase correctly (only if no provider error returns).
func testingManagerAll(asserts *assert.Assertions, managerStruct *manager, mockProvider *MockInterface) {
	// get all data should increase the statistic hit counter +1.
	mItem := &MockItem{}
	mockProvider.On("All").Once().Return([]Item{mItem, mItem, mItem}, nil)
	items, err := managerStruct.All()
	asserts.NoError(err)
	asserts.True(len(items) == 3)
	// check if all (foo, john, john2) statistics are increased +1.
	expectedStatisticsMap := map[string]counter{"testing_foo": {hit: 1, miss: 0, exists: true}, "names_john": {hit: 1, miss: 0, exists: true}, "names_john2": {hit: 1, miss: 0, exists: true}}
	asserts.True(fmt.Sprint(expectedStatisticsMap) == fmt.Sprint(managerStruct.statistics))

	// get all data should increase the statistic hit counter +1.
	mockProvider.On("All").Once().Return([]Item{mItem, mItem, mItem}, nil)
	items, err = managerStruct.All()
	asserts.NoError(err)
	asserts.True(len(items) == 3)
	// check if all (foo, john, john2) statistics are increased +1.
	expectedStatisticsMap = map[string]counter{"testing_foo": {hit: 2, miss: 0, exists: true}, "names_john": {hit: 2, miss: 0, exists: true}, "names_john2": {hit: 2, miss: 0, exists: true}}
	asserts.True(fmt.Sprint(expectedStatisticsMap) == fmt.Sprint(managerStruct.statistics))

	// get all data should not increase the statistic because the provider returns an error.
	mockProvider.On("All").Once().Return(nil, errors.New("some error"))
	items, err = managerStruct.All()
	asserts.Nil(items)
	asserts.Error(err)
	asserts.Equal([]Item(nil), items)
	// check if all (foo, john, john2) statistics are still the same.
	asserts.True(fmt.Sprint(expectedStatisticsMap) == fmt.Sprint(managerStruct.statistics))
}

// testingManagerGet is testing:
// - if the statistics increases correctly.
// - statistics only increase if the provider returns no error.
// - provider return value, errors.
// - statistics miss counter on non existing keys.
func testingManagerGet(asserts *assert.Assertions, managerStruct *manager, mockProvider *MockInterface) {

	// get with default prefix
	expectedItem := &MockItem{}
	mockProvider.On("Get", managerStruct.prefixedName(DefaultPrefix, "foo")).Once().Return(expectedItem, nil)
	item, err := managerStruct.Get(DefaultPrefix, "foo")
	asserts.NoError(err)
	asserts.Equal(expectedItem, item)

	// get with custom prefix
	mockProvider.On("Get", managerStruct.prefixedName("names", "john")).Once().Return(expectedItem, nil)
	item, err = managerStruct.Get("names", "john")
	asserts.Equal(expectedItem, item)

	// check if (foo, john) statistics are increased +1
	expectedStatisticsMap := map[string]counter{"testing_foo": {hit: 3, miss: 0, exists: true}, "names_john": {hit: 3, miss: 0, exists: true}, "names_john2": {hit: 2, miss: 0, exists: true}}
	asserts.True(fmt.Sprint(expectedStatisticsMap) == fmt.Sprint(managerStruct.statistics))

	// get with prefix again
	mockProvider.On("Get", managerStruct.prefixedName("names", "john")).Once().Return(expectedItem, nil)
	item, err = managerStruct.Get("names", "john")
	asserts.Equal(expectedItem, item)
	// check if only (john) statistics are increased +1
	expectedStatisticsMap = map[string]counter{"testing_foo": {hit: 3, miss: 0, exists: true}, "names_john": {hit: 4, miss: 0, exists: true}, "names_john2": {hit: 2, miss: 0, exists: true}}
	asserts.True(fmt.Sprint(expectedStatisticsMap) == fmt.Sprint(managerStruct.statistics))

	// testing an none existing entry
	mockProvider.On("Get", managerStruct.prefixedName("names", "john3")).Once().Return(nil, errors.New("not existing"))
	item, err = managerStruct.Get("names", "john3")
	asserts.Error(err)
	asserts.Nil(item)
	// check if all statistics are the same, only john3 should have been created with a miss hit.
	expectedStatisticsMap = map[string]counter{"testing_foo": {hit: 3, miss: 0, exists: true}, "names_john": {hit: 4, miss: 0, exists: true}, "names_john2": {hit: 2, miss: 0, exists: true}, "names_john3": {hit: 0, miss: 1, exists: false}}
	asserts.True(fmt.Sprint(expectedStatisticsMap) == fmt.Sprint(managerStruct.statistics))

	// testing that the not existing key is not getting increased if All() is called.
	mockProvider.On("All").Once().Return(nil, nil)
	items, err := managerStruct.All()
	asserts.Nil(items)
	asserts.NoError(err)
	expectedStatisticsMap = map[string]counter{"testing_foo": {hit: 4, miss: 0, exists: true}, "names_john": {hit: 5, miss: 0, exists: true}, "names_john2": {hit: 3, miss: 0, exists: true}, "names_john3": {hit: 0, miss: 1, exists: false}}
	asserts.True(fmt.Sprint(expectedStatisticsMap) == fmt.Sprint(managerStruct.statistics))

	// setting the none existing key (john3) from before.
	mockProvider.On("Set", managerStruct.prefixedName("names", "john3"), "doe", managerStruct.defaultExpiration).Once().Return(nil)
	err = managerStruct.Set("names", "john3", "doe", DefaultExpiration)
	expectedStatisticsMap = map[string]counter{"testing_foo": {hit: 4, miss: 0, exists: true}, "names_john": {hit: 5, miss: 0, exists: true}, "names_john2": {hit: 3, miss: 0, exists: true}, "names_john3": {hit: 0, miss: 1, exists: true}}
	asserts.True(fmt.Sprint(expectedStatisticsMap) == fmt.Sprint(managerStruct.statistics))
	asserts.NoError(err)

	// get john3 again, statistics should not be reset because of the miss counter and should be increased by +1.
	mockProvider.On("Get", managerStruct.prefixedName("names", "john3")).Once().Return(nil, nil)
	item, err = managerStruct.Get("names", "john3")
	asserts.NoError(err)
	asserts.Nil(item)
	// check if john3 statistics are increased +1
	expectedStatisticsMap = map[string]counter{"testing_foo": {hit: 4, miss: 0, exists: true}, "names_john": {hit: 5, miss: 0, exists: true}, "names_john2": {hit: 3, miss: 0, exists: true}, "names_john3": {hit: 1, miss: 1, exists: true}}
	asserts.True(fmt.Sprint(expectedStatisticsMap) == fmt.Sprint(managerStruct.statistics))
}

// testingManagerExist is testing:
// - if the provider return value is passed by.
// - statistics increases correctly
func testingManagerExist(asserts *assert.Assertions, managerStruct *manager, mockProvider *MockInterface) {
	mockProvider.On("Get", managerStruct.prefixedName(DefaultPrefix, "foo")).Once().Return(nil, nil)
	exists := managerStruct.Exist(DefaultPrefix, "foo")
	asserts.True(exists)
	// check if (foo) statistics are increased +1
	expectedStatisticsMap := map[string]counter{"testing_foo": {hit: 5, miss: 0, exists: true}, "names_john": {hit: 5, miss: 0, exists: true}, "names_john2": {hit: 3, miss: 0, exists: true}, "names_john3": {hit: 1, miss: 1, exists: true}}
	asserts.True(fmt.Sprint(expectedStatisticsMap) == fmt.Sprint(managerStruct.statistics))

	mockProvider.On("Get", managerStruct.prefixedName(DefaultPrefix, "notExisting")).Once().Return(nil, errors.New("not existing"))
	exists = managerStruct.Exist(DefaultPrefix, "notExisting")
	asserts.False(exists)
	// check if (notExisting) miss hit increased +1
	expectedStatisticsMap = map[string]counter{"testing_foo": {hit: 5, miss: 0, exists: true}, "names_john": {hit: 5, miss: 0, exists: true}, "names_john2": {hit: 3, miss: 0, exists: true}, "names_john3": {hit: 1, miss: 1, exists: true}, "testing_notExisting": {hit: 0, miss: 1, exists: false}}
	asserts.True(fmt.Sprint(expectedStatisticsMap) == fmt.Sprint(managerStruct.statistics))
}

// testingManagerPrefix is testing:
// - a prefix makes the correct provider calls.
// - increases the statistic correctly.
// - throws an error if the prefix does not exist.
// - throws an error if the prefixed name does not exist on provider side anymore.
func testingManagerPrefix(asserts *assert.Assertions, managerStruct *manager, mockProvider *MockInterface) {

	// getting all default prefixes
	mockProvider.On("Get", managerStruct.prefixedName(DefaultPrefix, "foo")).Once().Return(nil, nil)
	items, err := managerStruct.Prefix(DefaultPrefix)
	asserts.NoError(err)
	asserts.True(len(items) == 1)
	// check if (foo) statistics are increased +1
	expectedStatisticsMap := map[string]counter{"testing_foo": {hit: 6, miss: 0, exists: true}, "names_john": {hit: 5, miss: 0, exists: true}, "names_john2": {hit: 3, miss: 0, exists: true}, "names_john3": {hit: 1, miss: 1, exists: true}, "testing_notExisting": {miss: 1}}
	asserts.True(fmt.Sprint(expectedStatisticsMap) == fmt.Sprint(managerStruct.statistics))

	// getting named prefix
	mockProvider.On("Get", managerStruct.prefixedName("names", "john")).Once().Return(nil, nil)
	mockProvider.On("Get", managerStruct.prefixedName("names", "john2")).Once().Return(nil, nil)
	mockProvider.On("Get", managerStruct.prefixedName("names", "john3")).Once().Return(nil, nil)
	items, err = managerStruct.Prefix("names")
	asserts.NoError(err)
	asserts.True(len(items) == 3)
	// check if all (john,john2,john3) statistics are increased +1
	expectedStatisticsMap = map[string]counter{"testing_foo": {hit: 6, miss: 0, exists: true}, "names_john": {hit: 6, miss: 0, exists: true}, "names_john2": {hit: 4, miss: 0, exists: true}, "names_john3": {hit: 2, miss: 1, exists: true}, "testing_notExisting": {miss: 1}}
	asserts.True(fmt.Sprint(expectedStatisticsMap) == fmt.Sprint(managerStruct.statistics))

	// none existing prefix
	items, err = managerStruct.Prefix("notExisting")
	asserts.Error(err)
	asserts.True(len(items) == 0)

	//manipulating prefix map, to call a none existing provider cache item - should only happen if something is wrong on provider side.
	mockProvider.On("Get", managerStruct.prefixedName("names", "john")).Once().Return(nil, nil)
	mockProvider.On("Get", managerStruct.prefixedName("names", "john2")).Once().Return(nil, nil)
	mockProvider.On("Get", managerStruct.prefixedName("names", "john3")).Once().Return(nil, nil)
	mockProvider.On("Get", managerStruct.prefixedName("names", "notExisting")).Once().Return(nil, errors.New("not existing"))
	managerStruct.prefixes["names"] = append(managerStruct.prefixes["names"], "notExisting")
	items, err = managerStruct.Prefix("names")
	asserts.Error(err)
	asserts.True(len(items) == 0) // foo
	delete(managerStruct.statistics, managerStruct.prefixedName("names", "notExisting"))
	managerStruct.prefixes["names"] = managerStruct.prefixes["names"][:len(managerStruct.prefixes["names"])-1]
	expectedStatisticsMap = map[string]counter{"testing_foo": {hit: 6, miss: 0, exists: true}, "names_john": {hit: 7, miss: 0, exists: true}, "names_john2": {hit: 5, miss: 0, exists: true}, "names_john3": {hit: 3, miss: 1, exists: true}, "testing_notExisting": {miss: 1}}
	asserts.True(fmt.Sprint(expectedStatisticsMap) == fmt.Sprint(managerStruct.statistics))
}

// testingManagerDelete is testing:
// - delete existing cache item
// - delete none-existing cache item
func testingManagerDelete(asserts *assert.Assertions, managerStruct *manager, mockProvider *MockInterface) {
	// delete a cache item
	mockProvider.On("Delete", managerStruct.prefixedName(DefaultPrefix, "foo")).Once().Return(nil)
	err := managerStruct.Delete(DefaultPrefix, "foo")
	asserts.NoError(err)
	// testing the prefix struct
	expectedPrefixMap := map[string][]string{"names": {"john", "john2", "john3"}}
	asserts.True(fmt.Sprint(expectedPrefixMap) == fmt.Sprint(managerStruct.prefixes))
	// check if (foo) is deleted in statistics
	expectedStatisticsMap := map[string]counter{"names_john": {hit: 7, miss: 0, exists: true}, "names_john2": {hit: 5, miss: 0, exists: true}, "names_john3": {hit: 3, miss: 1, exists: true}, "testing_notExisting": {miss: 1}}
	asserts.True(fmt.Sprint(expectedStatisticsMap) == fmt.Sprint(managerStruct.statistics))

	// calling it again with the non existing name , error should return and the prefixes and statistics should be the same.
	mockProvider.On("Delete", managerStruct.prefixedName(DefaultPrefix, "foo")).Once().Return(errors.New("not existing"))
	err = managerStruct.Delete(DefaultPrefix, "foo")
	asserts.Error(err)
	asserts.True(fmt.Sprint(expectedPrefixMap) == fmt.Sprint(managerStruct.prefixes))
	asserts.True(fmt.Sprint(expectedStatisticsMap) == fmt.Sprint(managerStruct.statistics))
}

// testingManagerDeletePrefix is testing:
// - none existing prefix
// - test provider error
func testingManagerDeletePrefix(asserts *assert.Assertions, managerStruct *manager, mockProvider *MockInterface) {
	// none existing prefix - it will not hit the provider Delete function
	err := managerStruct.DeletePrefix(DefaultPrefix)
	asserts.Error(err)
	// testing if the prefix struct is still the same
	expectedPrefixMap := map[string][]string{"names": {"john", "john2", "john3"}}
	asserts.True(fmt.Sprint(expectedPrefixMap) == fmt.Sprint(managerStruct.prefixes))
	// check if all statistic is still the same
	expectedStatisticsMap := map[string]counter{"names_john": {hit: 7, miss: 0, exists: true}, "names_john2": {hit: 5, miss: 0, exists: true}, "names_john3": {hit: 3, miss: 1, exists: true}, "testing_notExisting": {miss: 1}}
	asserts.True(fmt.Sprint(expectedStatisticsMap) == fmt.Sprint(managerStruct.statistics))

	// test if provider returns an error, prefix and statistic should stay the same.
	mockProvider.On("Delete", managerStruct.prefixedName("names", "john")).Once().Return(errors.New("an error"))
	err = managerStruct.DeletePrefix("names")
	asserts.Error(err)
	asserts.True(fmt.Sprint(expectedPrefixMap) == fmt.Sprint(managerStruct.prefixes))
	asserts.True(fmt.Sprint(expectedStatisticsMap) == fmt.Sprint(managerStruct.statistics))

	// test existing entries.
	mockProvider.On("Delete", managerStruct.prefixedName("names", "john")).Once().Return(nil)
	mockProvider.On("Delete", managerStruct.prefixedName("names", "john2")).Once().Return(nil)
	mockProvider.On("Delete", managerStruct.prefixedName("names", "john3")).Once().Return(nil)
	err = managerStruct.DeletePrefix("names")
	asserts.NoError(err)
	// testing the prefix struct
	asserts.True(len(managerStruct.prefixes) == 0)
	// check if (john, john2, john3) is deleted in statistics
	expectedStatisticsMap = map[string]counter{"testing_notExisting": {miss: 1}}
	asserts.True(fmt.Sprint(expectedStatisticsMap) == fmt.Sprint(managerStruct.statistics))

	// testing reset statistic
	managerStruct.resetStatistic()
	asserts.True(len(managerStruct.statistics) == 0)
}

// testingManagerDeleteAll ist testing:
// - error return on provider side.
// - if all entries are deleted.
func testingManagerDeleteAll(asserts *assert.Assertions, managerStruct *manager, mockProvider *MockInterface) {
	// adding some data.
	testingManagerSet(asserts, managerStruct, mockProvider)
	// testing the prefix struct
	expectedPrefixMap := map[string][]string{"": {"foo"}, "names": {"john", "john2"}}
	asserts.True(fmt.Sprint(expectedPrefixMap) == fmt.Sprint(managerStruct.prefixes))
	// check if all (foo, john) statistics are increased +1
	expectedStatisticsMap := map[string]counter{"testing_foo": {hit: 0, miss: 0, exists: true}, "names_john": {hit: 0, miss: 0, exists: true}, "names_john2": {hit: 0, miss: 0, exists: true}}
	asserts.True(fmt.Sprint(expectedStatisticsMap) == fmt.Sprint(managerStruct.statistics))

	// testing error provider return
	mockProvider.On("DeleteAll").Once().Return(errors.New("an error"))
	err := managerStruct.DeleteAll()
	asserts.Error(err)
	asserts.True(fmt.Sprint(expectedPrefixMap) == fmt.Sprint(managerStruct.prefixes))
	asserts.True(fmt.Sprint(expectedStatisticsMap) == fmt.Sprint(managerStruct.statistics))

	// testing error provider return
	mockProvider.On("DeleteAll").Once().Return(nil)
	err = managerStruct.DeleteAll()
	asserts.NoError(err)
	// testing the prefix struct
	asserts.True(len(managerStruct.prefixes) == 0)
	asserts.True(len(managerStruct.statistics) == 0)
}

// Interface is an autogenerated mock type for the Interface type
type MockInterface struct {
	mock.Mock
}

// All provides a mock function with given fields:
func (_m *MockInterface) All() ([]Item, error) {
	ret := _m.Called()

	var r0 []Item
	if rf, ok := ret.Get(0).(func() []Item); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]Item)
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
func (_m *MockInterface) Delete(name string) error {
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
func (_m *MockInterface) DeleteAll() error {
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
func (_m *MockInterface) GC() {
	_m.Called()
}

// Get provides a mock function with given fields: name
func (_m *MockInterface) Get(name string) (Item, error) {
	ret := _m.Called(name)

	var r0 Item
	if rf, ok := ret.Get(0).(func(string) Item); ok {
		r0 = rf(name)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(Item)
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
func (_m *MockInterface) Set(name string, value interface{}, exp time.Duration) error {
	ret := _m.Called(name, value, exp)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, interface{}, time.Duration) error); ok {
		r0 = rf(name, value, exp)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Item is an autogenerated mock type for the Item type
type MockItem struct {
	mock.Mock
}

// Created provides a mock function with given fields:
func (_m *MockItem) Created() time.Time {
	ret := _m.Called()

	var r0 time.Time
	if rf, ok := ret.Get(0).(func() time.Time); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(time.Time)
	}

	return r0
}

// Expiration provides a mock function with given fields:
func (_m *MockItem) Expiration() time.Duration {
	ret := _m.Called()

	var r0 time.Duration
	if rf, ok := ret.Get(0).(func() time.Duration); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(time.Duration)
	}

	return r0
}

// Name provides a mock function with given fields:
func (_m *MockItem) Name() string {
	ret := _m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// Value provides a mock function with given fields:
func (_m *MockItem) Value() interface{} {
	ret := _m.Called()

	var r0 interface{}
	if rf, ok := ret.Get(0).(func() interface{}); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(interface{})
		}
	}

	return r0
}
