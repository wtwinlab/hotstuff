// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/relab/hotstuff (interfaces: Consensus)

// Package mocks is a generated GoMock package.
package mocks

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	hotstuff "github.com/relab/hotstuff"
)

// MockConsensus is a mock of Consensus interface.
type MockConsensus struct {
	ctrl     *gomock.Controller
	recorder *MockConsensusMockRecorder
}

// MockConsensusMockRecorder is the mock recorder for MockConsensus.
type MockConsensusMockRecorder struct {
	mock *MockConsensus
}

// NewMockConsensus creates a new mock instance.
func NewMockConsensus(ctrl *gomock.Controller) *MockConsensus {
	mock := &MockConsensus{ctrl: ctrl}
	mock.recorder = &MockConsensusMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockConsensus) EXPECT() *MockConsensusMockRecorder {
	return m.recorder
}

// IncreaseLastVotedView mocks base method.
func (m *MockConsensus) IncreaseLastVotedView(arg0 hotstuff.View) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "IncreaseLastVotedView", arg0)
}

// IncreaseLastVotedView indicates an expected call of IncreaseLastVotedView.
func (mr *MockConsensusMockRecorder) IncreaseLastVotedView(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "IncreaseLastVotedView", reflect.TypeOf((*MockConsensus)(nil).IncreaseLastVotedView), arg0)
}

// LastVote mocks base method.
func (m *MockConsensus) LastVote() hotstuff.View {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "LastVote")
	ret0, _ := ret[0].(hotstuff.View)
	return ret0
}

// LastVote indicates an expected call of LastVote.
func (mr *MockConsensusMockRecorder) LastVote() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "LastVote", reflect.TypeOf((*MockConsensus)(nil).LastVote))
}

// OnPropose mocks base method.
func (m *MockConsensus) OnPropose(arg0 hotstuff.ProposeMsg) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "OnPropose", arg0)
}

// OnPropose indicates an expected call of OnPropose.
func (mr *MockConsensusMockRecorder) OnPropose(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "OnPropose", reflect.TypeOf((*MockConsensus)(nil).OnPropose), arg0)
}

// OnVote mocks base method.
func (m *MockConsensus) OnVote(arg0 hotstuff.VoteMsg) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "OnVote", arg0)
}

// OnVote indicates an expected call of OnVote.
func (mr *MockConsensusMockRecorder) OnVote(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "OnVote", reflect.TypeOf((*MockConsensus)(nil).OnVote), arg0)
}

// Propose mocks base method.
func (m *MockConsensus) Propose() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Propose")
}

// Propose indicates an expected call of Propose.
func (mr *MockConsensusMockRecorder) Propose() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Propose", reflect.TypeOf((*MockConsensus)(nil).Propose))
}
