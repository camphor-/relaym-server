package entity

import (
	"sync"
	"time"
)

// SyncCheckTimer はSpotifyとの同期チェック用のタイマーです。タイマーが止まったことを確認するためのstopチャネルがあります。
// ref : http://okzk.hatenablog.com/entry/2015/12/01/001924
type SyncCheckTimer struct {
	timer  *time.Timer
	stopCh chan struct{}
}

// ExpireCh は指定設定された秒数経過したことを送るチャネルを返します。
func (s *SyncCheckTimer) ExpireCh() <-chan time.Time {
	return s.timer.C
}

// StopCh はタイマーがストップされたことを送るチャネルを返します。
func (s *SyncCheckTimer) StopCh() <-chan struct{} {
	return s.stopCh
}

func newSyncCheckTimer(d time.Duration) *SyncCheckTimer {
	return &SyncCheckTimer{
		timer:  time.NewTimer(d),
		stopCh: make(chan struct{}, 2),
	}
}

// SyncCheckTimerManager はSpotifyとの同期チェック用のタイマーを一括して管理する構造体です。
type SyncCheckTimerManager struct {
	timers map[string]*SyncCheckTimer
	mu     sync.Mutex
}

// NewSyncCheckTimerManager はSyncCheckTimerManagerのポインタを生成します。
func NewSyncCheckTimerManager() *SyncCheckTimerManager {
	return &SyncCheckTimerManager{
		timers: map[string]*SyncCheckTimer{},
	}
}

// CreateTimer は与えられたセッションの同期チェック用のタイマーを作成します。
// 既存のタイマーが存在する場合はstopしてから新しいタイマーを作成します。
func (m *SyncCheckTimerManager) CreateTimer(sessionID string, d time.Duration) *SyncCheckTimer {
	m.mu.Lock()
	defer m.mu.Unlock()

	if existing, ok := m.timers[sessionID]; ok {
		// 本来ならStopのGoDocコメントにある通り、<-t.Cとして、チャネルが空になっていることを確認すべきだが、
		// ExpireCh()の呼び出し側で受け取っているので問題ない。
		existing.timer.Stop()
		close(existing.stopCh)
	}
	timer := newSyncCheckTimer(d)
	m.timers[sessionID] = timer
	return timer
}

// StopTimer は与えられたセッションのタイマーを終了します。
func (m *SyncCheckTimerManager) StopTimer(sessionID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if timer, ok := m.timers[sessionID]; ok {
		if !timer.timer.Stop() {
			<-timer.timer.C
		}
		close(timer.stopCh)
		delete(m.timers, sessionID)
	}
}

// DeleteTimer は与えられたセッションのタイマーを削除します。
// StopTimerと異なり、すでにタイマーがExpireしてるときしか使えません。
func (m *SyncCheckTimerManager) DeleteTimer(sessionID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if timer, ok := m.timers[sessionID]; ok {
		close(timer.stopCh)
		delete(m.timers, sessionID)
	}
}

// GetTimer は与えられたセッションのタイマーを取得します。存在しない場合はfalseが返ります。
func (m *SyncCheckTimerManager) GetTimer(sessionID string) (*SyncCheckTimer, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if existing, ok := m.timers[sessionID]; ok {
		return existing, true
	}
	return nil, false
}
