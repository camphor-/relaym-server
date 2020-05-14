// Code generated by MockGen. DO NOT EDIT.
// Source: player.go

// Package mock_spotify is a generated GoMock package.
package mock_spotify

import (
	context "context"
	gomock "github.com/golang/mock/gomock"
	reflect "reflect"
)

// MockPlayer is a mock of Player interface
type MockPlayer struct {
	ctrl     *gomock.Controller
	recorder *MockPlayerMockRecorder
}

// MockPlayerMockRecorder is the mock recorder for MockPlayer
type MockPlayerMockRecorder struct {
	mock *MockPlayer
}

// NewMockPlayer creates a new mock instance
func NewMockPlayer(ctrl *gomock.Controller) *MockPlayer {
	mock := &MockPlayer{ctrl: ctrl}
	mock.recorder = &MockPlayerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockPlayer) EXPECT() *MockPlayerMockRecorder {
	return m.recorder
}

// CurrentlyPlaying mocks base method
func (m *MockPlayer) CurrentlyPlaying(ctx context.Context) (bool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CurrentlyPlaying", ctx)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CurrentlyPlaying indicates an expected call of CurrentlyPlaying
func (mr *MockPlayerMockRecorder) CurrentlyPlaying(ctx interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CurrentlyPlaying", reflect.TypeOf((*MockPlayer)(nil).CurrentlyPlaying), ctx)
}

// Play mocks base method
func (m *MockPlayer) Play(ctx context.Context) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Play", ctx)
	ret0, _ := ret[0].(error)
	return ret0
}

// Play indicates an expected call of Play
func (mr *MockPlayerMockRecorder) Play(ctx interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Play", reflect.TypeOf((*MockPlayer)(nil).Play), ctx)
}

// Pause mocks base method
func (m *MockPlayer) Pause(ctx context.Context) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Pause", ctx)
	ret0, _ := ret[0].(error)
	return ret0
}

// Pause indicates an expected call of Pause
func (mr *MockPlayerMockRecorder) Pause(ctx interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Pause", reflect.TypeOf((*MockPlayer)(nil).Pause), ctx)
}

// AddToQueue mocks base method
func (m *MockPlayer) AddToQueue(ctx context.Context, trackID string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "AddToQueue", ctx, trackID)
	ret0, _ := ret[0].(error)
	return ret0
}

// AddToQueue indicates an expected call of AddToQueue
func (mr *MockPlayerMockRecorder) AddToQueue(ctx, trackID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AddToQueue", reflect.TypeOf((*MockPlayer)(nil).AddToQueue), ctx, trackID)
}

// SetRepeatMode mocks base method
func (m *MockPlayer) SetRepeatMode(ctx context.Context, on bool) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SetRepeatMode", ctx, on)
	ret0, _ := ret[0].(error)
	return ret0
}

// SetRepeatMode indicates an expected call of SetRepeatMode
func (mr *MockPlayerMockRecorder) SetRepeatMode(ctx, on interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetRepeatMode", reflect.TypeOf((*MockPlayer)(nil).SetRepeatMode), ctx, on)
}
