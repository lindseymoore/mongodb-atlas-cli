// Code generated by MockGen. DO NOT EDIT.
// Source: internal/store/agents.go

// Package mocks is a generated GoMock package.
package mocks

import (
	gomock "github.com/golang/mock/gomock"
	opsmngr "github.com/mongodb/go-client-mongodb-ops-manager/opsmngr"
	reflect "reflect"
)

// MockAgentLister is a mock of AgentLister interface
type MockAgentLister struct {
	ctrl     *gomock.Controller
	recorder *MockAgentListerMockRecorder
}

// MockAgentListerMockRecorder is the mock recorder for MockAgentLister
type MockAgentListerMockRecorder struct {
	mock *MockAgentLister
}

// NewMockAgentLister creates a new mock instance
func NewMockAgentLister(ctrl *gomock.Controller) *MockAgentLister {
	mock := &MockAgentLister{ctrl: ctrl}
	mock.recorder = &MockAgentListerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockAgentLister) EXPECT() *MockAgentListerMockRecorder {
	return m.recorder
}

// Agents mocks base method
func (m *MockAgentLister) Agents(arg0, arg1 string) (*opsmngr.Agents, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Agents", arg0, arg1)
	ret0, _ := ret[0].(*opsmngr.Agents)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Agents indicates an expected call of Agents
func (mr *MockAgentListerMockRecorder) Agents(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Agents", reflect.TypeOf((*MockAgentLister)(nil).Agents), arg0, arg1)
}