// Code generated by mockery v2.14.0. DO NOT EDIT.

package apm

import (
	statsd "github.com/DataDog/datadog-go/statsd"
	mock "github.com/stretchr/testify/mock"

	time "time"
)

// MockStatsDClientInterface is an autogenerated mock type for the StatsDClientInterface type
type MockStatsDClientInterface struct {
	mock.Mock
}

// Close provides a mock function with given fields:
func (_m *MockStatsDClientInterface) Close() error {
	ret := _m.Called()

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Count provides a mock function with given fields: name, value, tags, rate
func (_m *MockStatsDClientInterface) Count(name string, value int64, tags []string, rate float64) error {
	ret := _m.Called(name, value, tags, rate)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, int64, []string, float64) error); ok {
		r0 = rf(name, value, tags, rate)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Decr provides a mock function with given fields: name, tags, rate
func (_m *MockStatsDClientInterface) Decr(name string, tags []string, rate float64) error {
	ret := _m.Called(name, tags, rate)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, []string, float64) error); ok {
		r0 = rf(name, tags, rate)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Distribution provides a mock function with given fields: name, value, tags, rate
func (_m *MockStatsDClientInterface) Distribution(name string, value float64, tags []string, rate float64) error {
	ret := _m.Called(name, value, tags, rate)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, float64, []string, float64) error); ok {
		r0 = rf(name, value, tags, rate)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Event provides a mock function with given fields: e
func (_m *MockStatsDClientInterface) Event(e *statsd.Event) error {
	ret := _m.Called(e)

	var r0 error
	if rf, ok := ret.Get(0).(func(*statsd.Event) error); ok {
		r0 = rf(e)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Flush provides a mock function with given fields:
func (_m *MockStatsDClientInterface) Flush() error {
	ret := _m.Called()

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Gauge provides a mock function with given fields: name, value, tags, rate
func (_m *MockStatsDClientInterface) Gauge(name string, value float64, tags []string, rate float64) error {
	ret := _m.Called(name, value, tags, rate)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, float64, []string, float64) error); ok {
		r0 = rf(name, value, tags, rate)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Histogram provides a mock function with given fields: name, value, tags, rate
func (_m *MockStatsDClientInterface) Histogram(name string, value float64, tags []string, rate float64) error {
	ret := _m.Called(name, value, tags, rate)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, float64, []string, float64) error); ok {
		r0 = rf(name, value, tags, rate)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Incr provides a mock function with given fields: name, tags, rate
func (_m *MockStatsDClientInterface) Incr(name string, tags []string, rate float64) error {
	ret := _m.Called(name, tags, rate)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, []string, float64) error); ok {
		r0 = rf(name, tags, rate)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// ServiceCheck provides a mock function with given fields: sc
func (_m *MockStatsDClientInterface) ServiceCheck(sc *statsd.ServiceCheck) error {
	ret := _m.Called(sc)

	var r0 error
	if rf, ok := ret.Get(0).(func(*statsd.ServiceCheck) error); ok {
		r0 = rf(sc)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Set provides a mock function with given fields: name, value, tags, rate
func (_m *MockStatsDClientInterface) Set(name string, value string, tags []string, rate float64) error {
	ret := _m.Called(name, value, tags, rate)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, string, []string, float64) error); ok {
		r0 = rf(name, value, tags, rate)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// SetWriteTimeout provides a mock function with given fields: d
func (_m *MockStatsDClientInterface) SetWriteTimeout(d time.Duration) error {
	ret := _m.Called(d)

	var r0 error
	if rf, ok := ret.Get(0).(func(time.Duration) error); ok {
		r0 = rf(d)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// SimpleEvent provides a mock function with given fields: title, text
func (_m *MockStatsDClientInterface) SimpleEvent(title string, text string) error {
	ret := _m.Called(title, text)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, string) error); ok {
		r0 = rf(title, text)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// SimpleServiceCheck provides a mock function with given fields: name, status
func (_m *MockStatsDClientInterface) SimpleServiceCheck(name string, status statsd.ServiceCheckStatus) error {
	ret := _m.Called(name, status)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, statsd.ServiceCheckStatus) error); ok {
		r0 = rf(name, status)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// TimeInMilliseconds provides a mock function with given fields: name, value, tags, rate
func (_m *MockStatsDClientInterface) TimeInMilliseconds(name string, value float64, tags []string, rate float64) error {
	ret := _m.Called(name, value, tags, rate)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, float64, []string, float64) error); ok {
		r0 = rf(name, value, tags, rate)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Timing provides a mock function with given fields: name, value, tags, rate
func (_m *MockStatsDClientInterface) Timing(name string, value time.Duration, tags []string, rate float64) error {
	ret := _m.Called(name, value, tags, rate)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, time.Duration, []string, float64) error); ok {
		r0 = rf(name, value, tags, rate)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

type mockConstructorTestingTNewMockStatsDClientInterface interface {
	mock.TestingT
	Cleanup(func())
}

// NewMockStatsDClientInterface creates a new instance of MockStatsDClientInterface. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewMockStatsDClientInterface(t mockConstructorTestingTNewMockStatsDClientInterface) *MockStatsDClientInterface {
	mock := &MockStatsDClientInterface{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
