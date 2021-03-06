// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/BarchDif/stm-like-api/internal/app/workerpool (interfaces: WorkerPool)

// Package mocks is a generated GoMock package.
package mocks

import (
	context "context"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
)

// MockWorkerPool is a mock of WorkerPool interface.
type MockWorkerPool struct {
	ctrl     *gomock.Controller
	recorder *MockWorkerPoolMockRecorder
}

// MockWorkerPoolMockRecorder is the mock recorder for MockWorkerPool.
type MockWorkerPoolMockRecorder struct {
	mock *MockWorkerPool
}

// NewMockWorkerPool creates a new mock instance.
func NewMockWorkerPool(ctrl *gomock.Controller) *MockWorkerPool {
	mock := &MockWorkerPool{ctrl: ctrl}
	mock.recorder = &MockWorkerPoolMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockWorkerPool) EXPECT() *MockWorkerPoolMockRecorder {
	return m.recorder
}

// Start mocks base method.
func (m *MockWorkerPool) Start(arg0 context.Context) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Start", arg0)
}

// Start indicates an expected call of Start.
func (mr *MockWorkerPoolMockRecorder) Start(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Start", reflect.TypeOf((*MockWorkerPool)(nil).Start), arg0)
}

// StopWait mocks base method.
func (m *MockWorkerPool) StopWait() chan struct{} {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "StopWait")
	ret0, _ := ret[0].(chan struct{})
	return ret0
}

// StopWait indicates an expected call of StopWait.
func (mr *MockWorkerPoolMockRecorder) StopWait() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "StopWait", reflect.TypeOf((*MockWorkerPool)(nil).StopWait))
}

// Submit mocks base method.
func (m *MockWorkerPool) Submit(arg0 func() error) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Submit", arg0)
}

// Submit indicates an expected call of Submit.
func (mr *MockWorkerPoolMockRecorder) Submit(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Submit", reflect.TypeOf((*MockWorkerPool)(nil).Submit), arg0)
}
