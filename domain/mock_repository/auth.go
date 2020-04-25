// Code generated by MockGen. DO NOT EDIT.
// Source: auth.go

// Package mock_repository is a generated GoMock package.
package mock_repository

import (
	entity "github.com/camphor-/relaym-server/domain/entity"
	gomock "github.com/golang/mock/gomock"
	oauth2 "golang.org/x/oauth2"
	reflect "reflect"
)

// MockAuth is a mock of Auth interface
type MockAuth struct {
	ctrl     *gomock.Controller
	recorder *MockAuthMockRecorder
}

// MockAuthMockRecorder is the mock recorder for MockAuth
type MockAuthMockRecorder struct {
	mock *MockAuth
}

// NewMockAuth creates a new mock instance
func NewMockAuth(ctrl *gomock.Controller) *MockAuth {
	mock := &MockAuth{ctrl: ctrl}
	mock.recorder = &MockAuthMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockAuth) EXPECT() *MockAuthMockRecorder {
	return m.recorder
}

// StoreORUpdateToken mocks base method
func (m *MockAuth) StoreORUpdateToken(userID string, token *oauth2.Token) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "StoreORUpdateToken", userID, token)
	ret0, _ := ret[0].(error)
	return ret0
}

// StoreORUpdateToken indicates an expected call of StoreORUpdateToken
func (mr *MockAuthMockRecorder) StoreORUpdateToken(userID, token interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "StoreORUpdateToken", reflect.TypeOf((*MockAuth)(nil).StoreORUpdateToken), userID, token)
}

// GetTokenByUserID mocks base method
func (m *MockAuth) GetTokenByUserID(userID string) (*oauth2.Token, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetTokenByUserID", userID)
	ret0, _ := ret[0].(*oauth2.Token)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetTokenByUserID indicates an expected call of GetTokenByUserID
func (mr *MockAuthMockRecorder) GetTokenByUserID(userID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetTokenByUserID", reflect.TypeOf((*MockAuth)(nil).GetTokenByUserID), userID)
}

// StoreSession mocks base method
func (m *MockAuth) StoreSession(sessionID, userID string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "StoreSession", sessionID, userID)
	ret0, _ := ret[0].(error)
	return ret0
}

// StoreSession indicates an expected call of StoreSession
func (mr *MockAuthMockRecorder) StoreSession(sessionID, userID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "StoreSession", reflect.TypeOf((*MockAuth)(nil).StoreSession), sessionID, userID)
}

// GetUserIDFromSession mocks base method
func (m *MockAuth) GetUserIDFromSession(sessionID string) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetUserIDFromSession", sessionID)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetUserIDFromSession indicates an expected call of GetUserIDFromSession
func (mr *MockAuthMockRecorder) GetUserIDFromSession(sessionID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetUserIDFromSession", reflect.TypeOf((*MockAuth)(nil).GetUserIDFromSession), sessionID)
}

// StoreState mocks base method
func (m *MockAuth) StoreState(authState *entity.AuthState) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "StoreState", authState)
	ret0, _ := ret[0].(error)
	return ret0
}

// StoreState indicates an expected call of StoreState
func (mr *MockAuthMockRecorder) StoreState(authState interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "StoreState", reflect.TypeOf((*MockAuth)(nil).StoreState), authState)
}

// FindStateByState mocks base method
func (m *MockAuth) FindStateByState(state string) (*entity.AuthState, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "FindStateByState", state)
	ret0, _ := ret[0].(*entity.AuthState)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// FindStateByState indicates an expected call of FindStateByState
func (mr *MockAuthMockRecorder) FindStateByState(state interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "FindStateByState", reflect.TypeOf((*MockAuth)(nil).FindStateByState), state)
}

// DeleteState mocks base method
func (m *MockAuth) DeleteState(state string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteState", state)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteState indicates an expected call of DeleteState
func (mr *MockAuthMockRecorder) DeleteState(state interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteState", reflect.TypeOf((*MockAuth)(nil).DeleteState), state)
}