// Code generated by MockGen. DO NOT EDIT.
// Source: session.go

// Package mock_repository is a generated GoMock package.
package mock_repository

import (
	entity "github.com/camphor-/relaym-server/domain/entity"
	gomock "github.com/golang/mock/gomock"
	oauth2 "golang.org/x/oauth2"
	reflect "reflect"
)

// MockSession is a mock of Session interface
type MockSession struct {
	ctrl     *gomock.Controller
	recorder *MockSessionMockRecorder
}

// MockSessionMockRecorder is the mock recorder for MockSession
type MockSessionMockRecorder struct {
	mock *MockSession
}

// NewMockSession creates a new mock instance
func NewMockSession(ctrl *gomock.Controller) *MockSession {
	mock := &MockSession{ctrl: ctrl}
	mock.recorder = &MockSessionMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockSession) EXPECT() *MockSessionMockRecorder {
	return m.recorder
}

// FindByID mocks base method
func (m *MockSession) FindByID(id string) (*entity.Session, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "FindByID", id)
	ret0, _ := ret[0].(*entity.Session)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// FindByID indicates an expected call of FindByID
func (mr *MockSessionMockRecorder) FindByID(id interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "FindByID", reflect.TypeOf((*MockSession)(nil).FindByID), id)
}

// StoreSession mocks base method
func (m *MockSession) StoreSession(arg0 *entity.Session) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "StoreSession", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// StoreSession indicates an expected call of StoreSession
func (mr *MockSessionMockRecorder) StoreSession(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "StoreSession", reflect.TypeOf((*MockSession)(nil).StoreSession), arg0)
}

// Update mocks base method
func (m *MockSession) Update(arg0 *entity.Session) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Update", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// Update indicates an expected call of Update
func (mr *MockSessionMockRecorder) Update(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Update", reflect.TypeOf((*MockSession)(nil).Update), arg0)
}

// StoreQueueTrack mocks base method
func (m *MockSession) StoreQueueTrack(arg0 *entity.QueueTrackToStore) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "StoreQueueTrack", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// StoreQueueTrack indicates an expected call of StoreQueueTrack
func (mr *MockSessionMockRecorder) StoreQueueTrack(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "StoreQueueTrack", reflect.TypeOf((*MockSession)(nil).StoreQueueTrack), arg0)
}

// FindCreatorTokenBySessionID mocks base method
func (m *MockSession) FindCreatorTokenBySessionID(arg0 string) (*oauth2.Token, string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "FindCreatorTokenBySessionID", arg0)
	ret0, _ := ret[0].(*oauth2.Token)
	ret1, _ := ret[1].(string)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// FindCreatorTokenBySessionID indicates an expected call of FindCreatorTokenBySessionID
func (mr *MockSessionMockRecorder) FindCreatorTokenBySessionID(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "FindCreatorTokenBySessionID", reflect.TypeOf((*MockSession)(nil).FindCreatorTokenBySessionID), arg0)
}

// ArchiveSessionsForBatch mocks base method
func (m *MockSession) ArchiveSessionsForBatch() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ArchiveSessionsForBatch")
	ret0, _ := ret[0].(error)
	return ret0
}

// ArchiveSessionsForBatch indicates an expected call of ArchiveSessionsForBatch
func (mr *MockSessionMockRecorder) ArchiveSessionsForBatch() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ArchiveSessionsForBatch", reflect.TypeOf((*MockSession)(nil).ArchiveSessionsForBatch))
}
