// Code generated by mockery v2.14.0. DO NOT EDIT.

package apm

import (
	time "time"

	mock "github.com/stretchr/testify/mock"
)

// MockMetrics is an autogenerated mock type for the Metrics type
type MockMetrics struct {
	mock.Mock
}

// Count provides a mock function with given fields: name, value, rate, tagPairs
func (_m *MockMetrics) Count(name string, value int64, rate float64, tagPairs ...string) {
	_va := make([]interface{}, len(tagPairs))
	for _i := range tagPairs {
		_va[_i] = tagPairs[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, name, value, rate)
	_ca = append(_ca, _va...)
	_m.Called(_ca...)
}

// Distribution provides a mock function with given fields: name, value, rate, tagPairs
func (_m *MockMetrics) Distribution(name string, value float64, rate float64, tagPairs ...string) {
	_va := make([]interface{}, len(tagPairs))
	for _i := range tagPairs {
		_va[_i] = tagPairs[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, name, value, rate)
	_ca = append(_ca, _va...)
	_m.Called(_ca...)
}

// Event provides a mock function with given fields: title, text, options
func (_m *MockMetrics) Event(title string, text string, options EventOptions) {
	_m.Called(title, text, options)
}

// Gauge provides a mock function with given fields: name, value, rate, tagPairs
func (_m *MockMetrics) Gauge(name string, value float64, rate float64, tagPairs ...string) {
	_va := make([]interface{}, len(tagPairs))
	for _i := range tagPairs {
		_va[_i] = tagPairs[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, name, value, rate)
	_ca = append(_ca, _va...)
	_m.Called(_ca...)
}

// Histogram provides a mock function with given fields: name, value, rate, tagPairs
func (_m *MockMetrics) Histogram(name string, value float64, rate float64, tagPairs ...string) {
	_va := make([]interface{}, len(tagPairs))
	for _i := range tagPairs {
		_va[_i] = tagPairs[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, name, value, rate)
	_ca = append(_ca, _va...)
	_m.Called(_ca...)
}

// Incr provides a mock function with given fields: name, rate, tagPairs
func (_m *MockMetrics) Incr(name string, rate float64, tagPairs ...string) {
	_va := make([]interface{}, len(tagPairs))
	for _i := range tagPairs {
		_va[_i] = tagPairs[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, name, rate)
	_ca = append(_ca, _va...)
	_m.Called(_ca...)
}

// Timing provides a mock function with given fields: name, value, rate, tagPairs
func (_m *MockMetrics) Timing(name string, value time.Duration, rate float64, tagPairs ...string) {
	_va := make([]interface{}, len(tagPairs))
	for _i := range tagPairs {
		_va[_i] = tagPairs[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, name, value, rate)
	_ca = append(_ca, _va...)
	_m.Called(_ca...)
}

type mockConstructorTestingTNewMockMetrics interface {
	mock.TestingT
	Cleanup(func())
}

// NewMockMetrics creates a new instance of MockMetrics. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewMockMetrics(t mockConstructorTestingTNewMockMetrics) *MockMetrics {
	mock := &MockMetrics{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
