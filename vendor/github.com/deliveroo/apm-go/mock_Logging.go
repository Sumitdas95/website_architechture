// Code generated by mockery v2.14.0. DO NOT EDIT.

package apm

import (
	mock "github.com/stretchr/testify/mock"
	zapcore "go.uber.org/zap/zapcore"
)

// MockLogging is an autogenerated mock type for the Logging type
type MockLogging struct {
	mock.Mock
}

// Debug provides a mock function with given fields: _a0, _a1
func (_m *MockLogging) Debug(_a0 string, _a1 ...zapcore.Field) {
	_va := make([]interface{}, len(_a1))
	for _i := range _a1 {
		_va[_i] = _a1[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, _a0)
	_ca = append(_ca, _va...)
	_m.Called(_ca...)
}

// Error provides a mock function with given fields: _a0, _a1
func (_m *MockLogging) Error(_a0 string, _a1 ...zapcore.Field) {
	_va := make([]interface{}, len(_a1))
	for _i := range _a1 {
		_va[_i] = _a1[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, _a0)
	_ca = append(_ca, _va...)
	_m.Called(_ca...)
}

// Fatal provides a mock function with given fields: _a0, _a1
func (_m *MockLogging) Fatal(_a0 string, _a1 ...zapcore.Field) {
	_va := make([]interface{}, len(_a1))
	for _i := range _a1 {
		_va[_i] = _a1[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, _a0)
	_ca = append(_ca, _va...)
	_m.Called(_ca...)
}

// Info provides a mock function with given fields: _a0, _a1
func (_m *MockLogging) Info(_a0 string, _a1 ...zapcore.Field) {
	_va := make([]interface{}, len(_a1))
	for _i := range _a1 {
		_va[_i] = _a1[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, _a0)
	_ca = append(_ca, _va...)
	_m.Called(_ca...)
}

// Panic provides a mock function with given fields: _a0, _a1
func (_m *MockLogging) Panic(_a0 string, _a1 ...zapcore.Field) {
	_va := make([]interface{}, len(_a1))
	for _i := range _a1 {
		_va[_i] = _a1[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, _a0)
	_ca = append(_ca, _va...)
	_m.Called(_ca...)
}

// Warn provides a mock function with given fields: _a0, _a1
func (_m *MockLogging) Warn(_a0 string, _a1 ...zapcore.Field) {
	_va := make([]interface{}, len(_a1))
	for _i := range _a1 {
		_va[_i] = _a1[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, _a0)
	_ca = append(_ca, _va...)
	_m.Called(_ca...)
}

type mockConstructorTestingTNewMockLogging interface {
	mock.TestingT
	Cleanup(func())
}

// NewMockLogging creates a new instance of MockLogging. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewMockLogging(t mockConstructorTestingTNewMockLogging) *MockLogging {
	mock := &MockLogging{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
